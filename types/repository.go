package types

import (
	"context"

	"github.com/wosai/deepmock/types/entity"
)

type (
	// RuleRepository Rule的存储对象
	RuleRepository interface {
		// CreateRule 创建规则
		CreateRule(context.Context, *entity.Rule) error
		// UpdateRule 更新规则，不管是Patch还是Put
		UpdateRule(context.Context, *entity.Rule) error
		// GetRule 根据ID获取规则
		GetRuleByID(context.Context, string) (*entity.Rule, error)
		// DeleteRule 物理删除规则
		DeleteRule(context.Context, string) error
		// Export 到处所有规则
		Export(context.Context) ([]*entity.Rule, error)
		// Import 导入所有规则
		Import(context.Context, ...*entity.Rule) error
	}

	MockRepository interface{}
)
