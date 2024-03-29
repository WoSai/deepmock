package application

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock/domain"
	"github.com/wosai/deepmock/misc"
	"github.com/wosai/deepmock/types"
	"go.uber.org/zap"
)

var (
	// MockApplication 全局的mockApplication对象
	MockApplication *mockApplication

	// ErrRuleNotFound 定义的无匹配规则时的错误
	ErrRuleNotFound = errors.New("rule not found")
)

type (
	// AsyncJob 异步的摆渡任务接口定义
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
		counter  uint64
	}
)

// BuildMockApplication mockApplication的工厂函数
func BuildMockApplication(rr domain.RuleRepository, er domain.ExecutorRepository, job AsyncJob) *mockApplication {
	MockApplication = &mockApplication{rule: rr, executor: er, job: job}
	go func() {
		job.WithRuleRepository(rr)
		job.WithExecutorRepository(er)
		t := time.NewTicker(job.Period())
		for range t.C {
			misc.Logger.Info("async job complete")
			if err := job.Do(); err != nil {
				misc.Logger.Error("occur error on job", zap.Error(err))
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
	}
	if rule.Weight != nil {
		r.Weight = make(map[string]domain.WeightFactor)
		for k, v := range rule.Weight {
			r.Weight[k] = v
		}
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
			RenderHeader:   reg.Template.RenderHeader,
			Header:         reg.Template.Header,
			HeaderTemplate: reg.Template.HeaderTemplate,
			Body:           reg.Template.Body,
			B64EncodedBody: reg.Template.B64EncodeBody,
			StatusCode:     reg.Template.StatusCode, // 默认不传，设置为200
		}
		if reg.Template.StatusCode == 0 {
			r.Template.StatusCode = http.StatusOK
		}
	}
	return r
}

func convertRuleEntity(rule *domain.Rule) *types.RuleDTO {
	r := &types.RuleDTO{
		ID:       rule.ID,
		Path:     rule.Path,
		Method:   rule.Method,
		Variable: rule.Variable,
	}
	if rule.Weight != nil {
		r.Weight = make(types.WeightDTO)
		for k, v := range rule.Weight {
			r.Weight[k] = v
		}
	}

	r.Regulations = make([]*types.RegulationDTO, len(rule.Regulations))
	for index, regulation := range rule.Regulations {
		r.Regulations[index] = convertRegulationVO(regulation)
	}
	return r
}

func convertRegulationVO(reg *domain.Regulation) *types.RegulationDTO {
	r := &types.RegulationDTO{
		IsDefault: reg.IsDefault,
		Template: &types.TemplateDTO{
			IsTemplate:     reg.Template.IsTemplate,
			RenderHeader:   reg.Template.RenderHeader,
			Header:         reg.Template.Header,
			HeaderTemplate: reg.Template.HeaderTemplate,
			StatusCode:     reg.Template.StatusCode,
			Body:           reg.Template.Body,
			B64EncodeBody:  reg.Template.B64EncodedBody,
		},
	}

	if reg.Filter != nil {
		r.Filter = &types.FilterDTO{
			Header: reg.Filter.Header,
			Query:  reg.Filter.Query,
			Body:   reg.Filter.Body,
		}
	}
	return r
}

// CreateRule 创建规则的user case
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

// GetRule 获取规则的user case
func (srv *mockApplication) GetRule(ctx context.Context, rid string) (*types.RuleDTO, error) {
	re, err := srv.rule.GetRuleByID(ctx, rid)
	if err != nil {
		misc.Logger.Error("failed to find rule record", zap.String("rule_id", rid), zap.Error(err))
		return nil, err
	}
	if err := re.Validate(); err != nil {
		misc.Logger.Error("failed to validate rule content", zap.String("rule_id", rid), zap.Error(err))
		return nil, err
	}
	rule := convertRuleEntity(re)
	return rule, nil
}

// DeleteRule 删除规则的user case
func (srv *mockApplication) DeleteRule(ctx context.Context, rid string) error {
	if err := srv.rule.DeleteRule(ctx, rid); err != nil {
		misc.Logger.Error("failed to delete rule entity", zap.String("rule_id", rid), zap.Error(err))
		return err
	}
	return nil
}

// PutRule 全量更新规则的user case
func (srv *mockApplication) PutRule(ctx context.Context, rule *types.RuleDTO) error {
	or, err := srv.rule.GetRuleByID(ctx, rule.ID)
	if err != nil {
		misc.Logger.Error("cannot found rule record with id", zap.String("rule_id", rule.ID), zap.Error(err))
		return err
	}

	nr := convertRuleDTO(rule)
	if err := or.Put(nr); err != nil {
		misc.Logger.Error("failed to validate rule after put", zap.String("rule_id", rule.ID), zap.Error(err))
		return err
	}
	if err := srv.rule.UpdateRule(ctx, or); err != nil {
		misc.Logger.Error("failed to update rule record", zap.String("rule_id", rule.ID), zap.Error(err))
		return err
	}
	misc.Logger.Info("update the rule record with id", zap.String("rule_id", rule.ID))
	return nil
}

// PatchRule 部分更新规则的user case
func (srv *mockApplication) PatchRule(ctx context.Context, rule *types.RuleDTO) error {
	or, err := srv.rule.GetRuleByID(ctx, rule.ID)
	if err != nil {
		misc.Logger.Error("cannot found rule record with id", zap.String("rule_id", rule.ID), zap.Error(err))
		return err
	}

	nr := convertRuleDTO(rule)
	if err := or.Patch(nr); err != nil {
		misc.Logger.Error("failed to validate rule after patch", zap.String("rule_id", rule.ID), zap.Error(err))
		return err
	}
	if err := srv.rule.UpdateRule(ctx, or); err != nil {
		misc.Logger.Error("failed to update rule record", zap.String("rule_id", rule.ID), zap.Error(err))
		return err
	}
	misc.Logger.Info("patch the rule record with id", zap.String("rule_id", rule.ID))
	return nil
}

// Export 导出的user case
func (srv *mockApplication) Export(ctx context.Context) ([]*types.RuleDTO, error) {
	res, err := srv.rule.Export(ctx)
	if err != nil {
		misc.Logger.Error("failed to export rules", zap.Error(err))
		return nil, err
	}
	rules := make([]*types.RuleDTO, len(res))
	for index, re := range res {
		if err := re.Validate(); err != nil {
			misc.Logger.Error("failed to convert as resource", zap.String("rule_id", re.ID), zap.Error(err))
			return nil, err
		}
		r := convertRuleEntity(re)
		rules[index] = r
	}
	return rules, nil
}

// Import 导入规则的user case
func (srv *mockApplication) Import(ctx context.Context, rules ...*types.RuleDTO) error {
	res := make([]*domain.Rule, len(rules))
	for index, rule := range rules {
		ru := convertRuleDTO(rule)
		if err := ru.Validate(); err != nil {
			misc.Logger.Error("failed to validate rule content", zap.String("rule_id", rule.ID), zap.Error(err))
			return err
		}

		res[index] = ru
	}

	if err := srv.rule.Import(ctx, res...); err != nil {
		misc.Logger.Error("failed to import rules", zap.Error(err))
		return err
	}
	return nil
}

// MockAPI Mock接口的user case
func (srv *mockApplication) MockAPI(ctx *fasthttp.RequestCtx) error {
	index := atomic.AddUint64(&srv.counter, 1)
	misc.Logger.Info("received request", zap.Uint64("index", index), zap.ByteString("path", ctx.Request.URI().Path()), zap.ByteString("method", ctx.Request.Header.Method()))
	exec, founded := srv.executor.FindExecutor(context.TODO(), ctx.Request.URI().Path(), ctx.Request.Header.Method())
	if !founded {
		misc.Logger.Warn("no matched rule founded", zap.Uint64("index", index))
		return ErrRuleNotFound
	}
	misc.Logger.Info("found matched rule", zap.Uint64("index", index), zap.String("rule_id", exec.ID))
	return exec.FindRegulationExecutor(&ctx.Request).Render(ctx, exec.Variable, exec.Weight.DiceAll())
}
