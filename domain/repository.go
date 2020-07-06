package domain

import "context"

type (
	RuleRepository interface {
		CreateRule(context.Context, *Rule) error
		UpdateRule(context.Context, *Rule) error
		GetRuleByID(context.Context, string) (*Rule, error)
		DeleteRule(context.Context, string) error
		Export(context.Context) ([]*Rule, error)
		Import(context.Context, ...*Rule) error
	}

	ExecutorRepository interface {
		GetExecutor(context.Context, *Executor) error
		Import(context.Context, ...*Executor) error
	}
)
