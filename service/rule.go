package service

import (
	"context"
	"errors"

	jsoniter "github.com/json-iterator/go"

	"github.com/wosai/deepmock"
	"go.uber.org/zap"

	"github.com/wosai/deepmock/types"
	"github.com/wosai/deepmock/types/entity"
	"github.com/wosai/deepmock/types/resource"
)

var (
	RuleService rule
	json        = jsoniter.ConfigCompatibleWithStandardLibrary
)

type (
	rule struct {
		repo types.RuleRepository
	}
)

func (srv rule) CreateRule(rule *resource.Rule) (string, error) {
	if err := ValidateRule(rule); err != nil {
		deepmock.Logger.Error("failed to validate rule content", zap.Error(err))
		return "", err
	}

	re, err := convertAsEntity(rule)
	if err != nil {
		deepmock.Logger.Error("failed to convert as an entity object", zap.Error(err))
	}
	err = srv.repo.CreateRule(context.Background(), re)
	if err != nil {
		deepmock.Logger.Error("failed to create rule record", zap.Error(err))
		return "", err
	}
	deepmock.Logger.Info("created new rule record with id", zap.String("rule_id", re.ID))
	return re.ID, nil
}

func (srv rule) GetRule(rid string) (*resource.Rule, error) {
	if rid == "" {
		deepmock.Logger.Error("missing rule id")
		return nil, errors.New("missing rule id")
	}
	re, err := srv.repo.GetRuleByID(context.Background(), rid)
	if err != nil {
		deepmock.Logger.Error("failed to find rule record", zap.String("rule_id", rid), zap.Error(err))
		return nil, err
	}
	rs, err := convertAsResource(re)
	if err != nil {
		deepmock.Logger.Error("failed to convert as resource", zap.String("rule_id", re.ID), zap.Error(err))
	}
	return rs, err
}

func (srv rule) DeleteRule(rid string) error {
	if rid == "" {
		return errors.New("missing rule id")
	}
	if err := srv.repo.DeleteRule(context.TODO(), rid); err != nil {
		deepmock.Logger.Error("failed to delete rule entity", zap.String("rule_id", rid), zap.Error(err))
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
		deepmock.Logger.Error("failed to validate rule entity", zap.String("rule_id", nr.ID), zap.Error(err))
		return err
	}

	re, err := convertAsEntity(nr)
	if err != nil {
		deepmock.Logger.Error("failed to convert as rule entity", zap.String("rule_id", nr.ID), zap.Error(err))
		return err
	}
	if err = srv.repo.UpdateRule(context.TODO(), re); err != nil {
		deepmock.Logger.Error("failed to update rule record", zap.String("rule_id", nr.ID), zap.Error(err))
		return err
	}
	deepmock.Logger.Info("replaced rule entity", zap.String("rule_id", nr.ID))
	return nil
}

func (srv rule) PatchRule(nr *resource.Rule) error {
	re, err := srv.repo.GetRuleByID(context.TODO(), nr.ID)
	if err != nil {
		return err
	}

	rs, err := convertAsResource(re)
	if err != nil {
		deepmock.Logger.Error("failed to convert as resource", zap.String("rule_id", re.ID), zap.Error(err))
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
		deepmock.Logger.Error("failed to validate rule", zap.String("rule_id", rs.ID), zap.Error(err))
		return err
	}

	re, err = convertAsEntity(rs)
	if err != nil {
		deepmock.Logger.Error("failed to convert as rule entity", zap.String("rule_id", rs.ID), zap.Error(err))
		return err
	}
	err = srv.repo.UpdateRule(context.TODO(), re)
	if err != nil {
		deepmock.Logger.Error("failed to patch rule entity", zap.String("rule_id", rs.ID), zap.Error(err))
	}
	return err
}

func (srv rule) Export() ([]*resource.Rule, error) {
	res, err := srv.repo.Export(context.TODO())
	if err != nil {
		deepmock.Logger.Error("failed to export rules", zap.Error(err))
		return nil, err
	}
	rules := make([]*resource.Rule, len(res))
	for index, re := range res {
		r, err := convertAsResource(re)
		if err != nil {
			deepmock.Logger.Error("failed to convert as resource", zap.String("rule_id", re.ID), zap.Error(err))
			return nil, err
		}
		rules[index] = r
	}
	return rules, nil
}

func (srv rule) Import(rules ...*resource.Rule) error {
	if len(rules) == 0 {
		deepmock.Logger.Error("disallowed to import empty rule")
		return errors.New("nothing to import")
	}
	res := make([]*entity.Rule, len(rules))
	for index, rule := range rules {
		if err := ValidateRule(rule); err != nil {
			deepmock.Logger.Error("failed to validate rule", zap.String("rule_id", rule.ID), zap.Error(err))
			return err
		}

		re, err := convertAsEntity(rule)
		if err != nil {
			deepmock.Logger.Error("failed to convert as entity", zap.String("rule_id", rule.ID), zap.Error(err))
			return err
		}

		res[index] = re
	}

	if err := srv.repo.Import(context.TODO(), res...); err != nil {
		deepmock.Logger.Error("failed to import rules", zap.Error(err))
		return err
	}
	return nil
}
