package application

import (
	"context"
	"errors"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/wosai/deepmock/domain"
	"github.com/wosai/deepmock/misc"
	"github.com/wosai/deepmock/types"
	"github.com/wosai/deepmock/types/entity"
	"github.com/wosai/deepmock/types/resource"
	"go.uber.org/zap"
)

var (
	MockApplication *mockApplication
	json            = jsoniter.ConfigCompatibleWithStandardLibrary
)

type (
	AsyncJob interface {
		Period() time.Duration
		Do() error
		WithRuleRepository(domain.RuleRepository)
		WithExecutorRepository(domain.ExecutorRepository)
	}

	mockApplication struct {
		rule     domain.RuleRepository
		executor domain.ExecutorRepository
		job      AsyncJob
	}
)

func BuildRuleService(rr domain.RuleRepository, er domain.ExecutorRepository, job AsyncJob) *mockApplication {
	MockApplication = &mockApplication{rule: rr, executor: er, job: job}
	go func() {
		job.WithRuleRepository(rr)
		job.WithExecutorRepository(er)
		t := time.NewTicker(job.Period())
		for {
			<-t.C
			if err := job.Do(); err != nil {
				misc.Logger.Error("occur error on job", zap.Error(err))
				t.Stop()
				break
			}
		}
	}()
	return MockApplication
}

func convertRuleDTO(rule *types.RuleDTO) *domain.Rule {
	r := &domain.Rule{
		ID:       rule.ID,
		Path:     rule.Path,
		Method:   rule.Method,
		Variable: rule.Variable,
		Weight:   rule.Weight,
	}
	r.Regulations = make([]*domain.Regulation, len(rule.Regulations))

	for index, regulation := range rule.Regulations {
		r.Regulations[index] = convertRegulationDTO(regulation)
	}
	return r
}

func convertRegulationDTO(reg *types.RegulationDTO) *domain.Regulation {
	r := &domain.Regulation{IsDefault: reg.IsDefault}
	if reg.Filter != nil {
		r.Filter = &domain.Filter{
			Query:  reg.Filter.Query,
			Header: reg.Filter.Header,
			Body:   reg.Filter.Body,
		}
	}
	if reg.Template != nil {
		r.Template = &domain.Template{
			IsTemplate:     reg.Template.IsTemplate,
			Header:         reg.Template.Header,
			Body:           reg.Template.Body,
			B64EncodedBody: reg.Template.B64EncodeBody,
		}
		if reg.Template.StatusCode == 0 {
			r.Template.StatusCode = http.StatusOK
		}
	}
	return r
}

func (srv *mockApplication) CreateRule(ctx context.Context, rule *types.RuleDTO) (string, error) {
	ru := convertRuleDTO(rule)
	rid, _ := ru.SupplyID()
	if err := ru.Validate(); err != nil {
		misc.Logger.Error("failed to validate rule content", zap.Error(err))
		return rid, err
	}

	if err := srv.rule.CreateRule(ctx, ru); err != nil {
		misc.Logger.Error("failed to create rule record", zap.Error(err))
		return rid, err
	}
	misc.Logger.Info("created new rule record with id", zap.String("rule_id", ru.ID))
	return rid, nil
}

func (srv *mockApplication) GetRule(rid string) (*resource.Rule, error) {
	if rid == "" {
		misc.Logger.Error("missing rule id")
		return nil, errors.New("missing rule id")
	}
	re, err := srv.repo.GetRuleByID(context.Background(), rid)
	if err != nil {
		misc.Logger.Error("failed to find rule record", zap.String("rule_id", rid), zap.Error(err))
		return nil, err
	}
	rs, err := convertAsResource(re)
	if err != nil {
		misc.Logger.Error("failed to convert as resource", zap.String("rule_id", re.ID), zap.Error(err))
	}
	return rs, err
}

func (srv rule) DeleteRule(rid string) error {
	if rid == "" {
		return errors.New("missing rule id")
	}
	if err := srv.repo.DeleteRule(context.TODO(), rid); err != nil {
		misc.Logger.Error("failed to delete rule entity", zap.String("rule_id", rid), zap.Error(err))
		return err
	}
	return nil
}

func (srv rule) PutRule(nr *resource.Rule) error {
	or, err := srv.GetRule(nr.ID)
	if err != nil {
		return err
	}
	nr.ID = or.ID
	nr.Path = or.Path
	nr.Method = or.Method
	nr.Version = or.Version
	if err := ValidateRule(nr); err != nil {
		misc.Logger.Error("failed to validate rule entity", zap.String("rule_id", nr.ID), zap.Error(err))
		return err
	}

	re, err := convertAsEntity(nr)
	if err != nil {
		misc.Logger.Error("failed to convert as rule entity", zap.String("rule_id", nr.ID), zap.Error(err))
		return err
	}
	if err = srv.repo.UpdateRule(context.TODO(), re); err != nil {
		misc.Logger.Error("failed to update rule record", zap.String("rule_id", nr.ID), zap.Error(err))
		return err
	}
	misc.Logger.Info("replaced rule entity", zap.String("rule_id", nr.ID))
	return nil
}

func (srv rule) PatchRule(nr *resource.Rule) error {
	re, err := srv.repo.GetRuleByID(context.TODO(), nr.ID)
	if err != nil {
		return err
	}

	rs, err := convertAsResource(re)
	if err != nil {
		misc.Logger.Error("failed to convert as resource", zap.String("rule_id", re.ID), zap.Error(err))
		return err
	}

	if nr.Variable != nil && len(nr.Variable) > 0 {
		m := make(resource.Variable)
		if rs.Variable != nil && len(rs.Variable) > 0 {
			m = rs.Variable
		}
		for k, v := range nr.Variable {
			m[k] = v
		}
		rs.Variable = m
	}

	if nr.Weight != nil && len(nr.Weight) > 0 {
		m := make(resource.Weight)
		if rs.Weight != nil && len(nr.Weight) > 0 {
			m = rs.Weight
		}
		for k, v := range nr.Weight {
			d, exist := m[k]
			if !exist {
				m[k] = v
			} else {
				for i, j := range v {
					d[i] = j
				}
			}
		}
		rs.Weight = m
	}

	if nr.Responses != nil && len(nr.Responses) > 0 {
		rs.Responses = nr.Responses
	}

	if err = ValidateRule(rs); err != nil {
		misc.Logger.Error("failed to validate rule", zap.String("rule_id", rs.ID), zap.Error(err))
		return err
	}

	re, err = convertAsEntity(rs)
	if err != nil {
		misc.Logger.Error("failed to convert as rule entity", zap.String("rule_id", rs.ID), zap.Error(err))
		return err
	}
	err = srv.repo.UpdateRule(context.TODO(), re)
	if err != nil {
		misc.Logger.Error("failed to patch rule entity", zap.String("rule_id", rs.ID), zap.Error(err))
	}
	return err
}

func (srv rule) Export() ([]*resource.Rule, error) {
	res, err := srv.repo.Export(context.TODO())
	if err != nil {
		misc.Logger.Error("failed to export rules", zap.Error(err))
		return nil, err
	}
	rules := make([]*resource.Rule, len(res))
	for index, re := range res {
		r, err := convertAsResource(re)
		if err != nil {
			misc.Logger.Error("failed to convert as resource", zap.String("rule_id", re.ID), zap.Error(err))
			return nil, err
		}
		rules[index] = r
	}
	return rules, nil
}

func (srv rule) Import(rules ...*resource.Rule) error {
	if len(rules) == 0 {
		misc.Logger.Error("disallowed to import empty rule")
		return errors.New("nothing to import")
	}
	res := make([]*entity.Rule, len(rules))
	for index, rule := range rules {
		if err := ValidateRule(rule); err != nil {
			misc.Logger.Error("failed to validate rule", zap.String("rule_id", rule.ID), zap.Error(err))
			return err
		}

		re, err := convertAsEntity(rule)
		if err != nil {
			misc.Logger.Error("failed to convert as entity", zap.String("rule_id", rule.ID), zap.Error(err))
			return err
		}

		res[index] = re
	}

	if err := srv.repo.Import(context.TODO(), res...); err != nil {
		misc.Logger.Error("failed to import rules", zap.Error(err))
		return err
	}
	return nil
}
