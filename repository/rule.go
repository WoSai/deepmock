package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/wosai/deepmock/types"

	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/wosai/deepmock/domain"
	"github.com/wosai/deepmock/misc"
	"github.com/wosai/deepmock/types/entity"
	"go.uber.org/zap"
)

type (
	RuleRepositoryImpl struct {
		db    *sql.DB
		table string
	}
)

func convertEntity(rule *domain.Rule) *types.RuleDo {
	return nil
}

func convertDataObject(rule *types.RuleDo) *domain.Rule {
	return nil
}

func NewRuleRepository(db *sql.DB) *RuleRepositoryImpl {
	return &RuleRepositoryImpl{db: db, table: "rule"}
}

func (r *RuleRepositoryImpl) CreateRule(ctx context.Context, rule *domain.Rule) error {
	record, err := scanner.Map(rule, "ddb")
	if err != nil {
		return err
	}
	record["version"] = 1
	record["disabled"] = 0
	delete(record, "ctime")
	delete(record, "mtime")
	query, values, err := builder.BuildInsert(r.table, []map[string]interface{}{record})
	if err != nil {
		return err
	}
	misc.Logger.Info(query, zap.Any("values", values))
	_, err = r.db.ExecContext(ctx, query, values...)
	return err
}

func (r *RuleRepositoryImpl) UpdateRule(ctx context.Context, rule *domain.Rule) error {
	cond, values, err := builder.BuildUpdate(
		r.table,
		map[string]interface{}{"id": rule.ID},
		map[string]interface{}{
			"variable":  rule.Variable,
			"weight":    rule.Weight,
			"responses": rule.Regulations,
			"version":   rule.Version + 1,
		},
	)
	if err != nil {
		return err
	}
	misc.Logger.Info(cond, zap.Any("values", values))
	_, err = r.db.ExecContext(ctx, cond, values...)
	return err
}

func (r *RuleRepositoryImpl) GetRuleByID(ctx context.Context, rid string) (*domain.Rule, error) {
	query, values, _ := builder.BuildSelect(
		r.table,
		map[string]interface{}{"id": rid, "_limit": []uint{1}},
		[]string{"*"},
	)
	misc.Logger.Info(query, zap.Any("values", values))
	rows, err := r.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var rules []*entity.Rule
	err = scanner.Scan(rows, &rules)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, errors.New("cannot find rule by id: " + rid)
	}
	return rules[0], nil
}

func (r *RuleRepositoryImpl) DeleteRule(ctx context.Context, rid string) error {
	cond, values, err := builder.BuildDelete(r.table, map[string]interface{}{"id": rid})
	if err != nil {
		return err
	}
	misc.Logger.Info(cond, zap.Any("values", values))
	_, err = r.db.ExecContext(ctx, cond, values...)
	return err
}

func (r *RuleRepositoryImpl) Export(ctx context.Context) ([]*domain.Rule, error) {
	query, values, _ := builder.BuildSelect(r.table, nil, []string{"*"})
	misc.Logger.Info(query, zap.Any("values", values))
	rows, err := r.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []*entity.Rule
	err = scanner.Scan(rows, &rules)
	return rules, err
}

func (r *RuleRepositoryImpl) Import(ctx context.Context, rules ...*domain.Rule) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	// 清空数据库
	cond, values, _ := builder.BuildDelete(r.table, nil)
	misc.Logger.Info(cond, zap.Any("values", values))
	_, err = tx.ExecContext(ctx, cond, values...)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	var records = make([]map[string]interface{}, len(rules))
	for i, rule := range rules {
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
	misc.Logger.Info(cond, zap.Any("values", values))
	_, err = tx.ExecContext(ctx, cond, values...)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
