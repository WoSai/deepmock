package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	jsoniter "github.com/json-iterator/go"
	"github.com/wosai/deepmock/domain"
	"github.com/wosai/deepmock/types"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type (
	RuleRepository struct {
		db    *sql.DB
		table string
	}
)

func convertRuleEntity(rule *domain.Rule) (*types.RuleDO, error) {
	do := &types.RuleDO{
		ID:       rule.ID,
		Path:     rule.Path,
		Method:   rule.Method,
		Version:  rule.Version,
		Disabled: false,
	}
	var err error
	if rule.Variable != nil {
		if do.Variable, err = json.Marshal(rule.Variable); err != nil {
			return nil, err
		}
	}
	if rule.Weight != nil {
		if do.Weight, err = json.Marshal(rule.Weight); err != nil {
			return nil, err
		}
	}
	if rule.Regulations != nil {
		if do.Responses, err = json.Marshal(rule.Regulations); err != nil {
			return nil, err
		}
	}
	return do, nil
}

// todo: 现在通过在entity上加tag实现转换，domain层不应该感知infra的数据结构，不合理，之后要优化
func convertRuleDO(rule *types.RuleDO) (*domain.Rule, error) {
	entity := &domain.Rule{
		ID:      rule.ID,
		Path:    rule.Path,
		Method:  rule.Method,
		Version: rule.Version,
	}
	if rule.Weight != nil {
		if err := json.Unmarshal(rule.Weight, &entity.Weight); err != nil {
			return nil, err
		}
	}

	if rule.Variable != nil {
		if err := json.Unmarshal(rule.Variable, &entity.Variable); err != nil {
			return nil, err
		}
	}

	if err := json.Unmarshal(rule.Responses, &entity.Regulations); err != nil {
		return nil, err
	}
	return entity, nil
}

func NewRuleRepository(db *sql.DB) *RuleRepository {
	return &RuleRepository{db: db, table: "rule"}
}

func (r *RuleRepository) CreateRule(ctx context.Context, rule *domain.Rule) error {
	do, err := convertRuleEntity(rule)
	if err != nil {
		return err
	}

	record, err := scanner.Map(do, "ddb")
	if err != nil {
		return err
	}
	delete(record, "ctime")
	delete(record, "mtime")
	query, values, err := builder.BuildInsert(r.table, []map[string]interface{}{record})
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, values...)
	return err
}

func (r *RuleRepository) UpdateRule(ctx context.Context, rule *domain.Rule) error {
	do, err := convertRuleEntity(rule)
	if err != nil {
		return err
	}

	cond, values, err := builder.BuildUpdate(
		r.table,
		map[string]interface{}{
			"id":      do.ID,
			"version": do.Version - 1,
		},
		map[string]interface{}{
			"variable":  do.Variable,
			"weight":    do.Weight,
			"responses": do.Responses,
			"version":   do.Version,
		},
	)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, cond, values...)
	return err
}

func (r *RuleRepository) GetRuleByID(ctx context.Context, rid string) (*domain.Rule, error) {
	query, values, _ := builder.BuildSelect(
		r.table,
		map[string]interface{}{
			"id":       rid,
			"_limit":   []uint{1},
			"disabled": 0,
		},
		[]string{"*"},
	)
	rows, err := r.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var rules []*types.RuleDO
	err = scanner.Scan(rows, &rules)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, errors.New("cannot find rule by id: " + rid)
	}

	return convertRuleDO(rules[0])
}

func (r *RuleRepository) DeleteRule(ctx context.Context, rid string) error {
	cond, values, err := builder.BuildDelete(r.table, map[string]interface{}{"id": rid, "disabled": 0})
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, cond, values...)
	return err
}

func (r *RuleRepository) Export(ctx context.Context) ([]*domain.Rule, error) {
	query, values, _ := builder.BuildSelect(
		r.table,
		map[string]interface{}{"disabled": 0},
		[]string{"*"},
	)
	rows, err := r.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var rules []*types.RuleDO
	err = scanner.Scan(rows, &rules)
	if err != nil {
		return nil, err
	}
	entities := make([]*domain.Rule, len(rules))
	for index, rule := range rules {
		entity, err := convertRuleDO(rule)
		if err != nil {
			return nil, err
		}
		entities[index] = entity
	}
	return entities, nil
}

func (r *RuleRepository) Import(ctx context.Context, rules ...*domain.Rule) error {
	dataObjects := make([]*types.RuleDO, len(rules))
	ids := make([]string, len(rules))
	for index, rule := range rules {
		data, err := convertRuleEntity(rule)
		if err != nil {
			return err
		}
		dataObjects[index] = data
		ids[index] = data.ID
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	// 清空存在的记录
	cond, values, _ := builder.BuildDelete(r.table, map[string]interface{}{
		"id in": ids,
	})
	_, err = tx.ExecContext(ctx, cond, values...)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	var records = make([]map[string]interface{}, len(dataObjects))
	for i, rule := range dataObjects {
		record, err := scanner.Map(rule, "ddb")
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		delete(record, "ctime")
		delete(record, "mtime")
		records[i] = record
	}

	cond, values, err = builder.BuildInsert(r.table, records)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.ExecContext(ctx, cond, values...)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
