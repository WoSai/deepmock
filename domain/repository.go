package domain

import "context"

type (
	// RuleRepository 规则存储库接口定义
	RuleRepository interface {
		CreateRule(context.Context, *Rule) error
		UpdateRule(context.Context, *Rule) error
		GetRuleByID(context.Context, string) (*Rule, error)
		DeleteRule(context.Context, string) error
		Export(context.Context) ([]*Rule, error)
		Import(context.Context, ...*Rule) error
	}

	// ExecutorRepository 执行器接口定义
	ExecutorRepository interface {
		FindExecutor(context.Context, []byte, []byte) (*Executor, bool)
		ImportAll(context.Context, ...*Executor)
	}
)
