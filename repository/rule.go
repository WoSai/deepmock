package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/wosai/deepmock/types"
	"github.com/wosai/deepmock/types/entity"
)

type (
	RuleStorage struct {
		db    *sql.DB
		table string
	}
)

func NewRuleRepository(db *sql.DB) types.RuleRepository {
	return &RuleStorage{db: db, table: "rule"}
}

func (r *RuleStorage) CreateRule(ctx context.Context, rule *entity.Rule) error {
	record, err := scanner.Map(rule, "ddb")
	if err != nil {
		return err
	}
	record["version"] = 0
	delete(record, "ctime")
	delete(record, "mtime")
	query, values, err := builder.BuildInsert(r.table, []map[string]interface{}{record})
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, values...)
	return err
}

func (r *RuleStorage) UpdateRule(ctx context.Context, rule *entity.Rule) error {
	cond, values, err := builder.BuildUpdate(
		r.table,
		map[string]interface{}{"id": rule.ID},
		map[string]interface{}{
			"variable":  rule.Variable,
			"weight":    rule.Weight,
			"responses": rule.Responses,
			"version":   rule.Version + 1,
		},
	)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, cond, values...)
	return err
}

func (r *RuleStorage) GetRuleByID(ctx context.Context, rid string) (*entity.Rule, error) {
	query, values, _ := builder.BuildSelect(
		r.table,
		map[string]interface{}{"id": rid, "_limit": []uint{1}},
		[]string{"*"},
	)
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

func (r *RuleStorage) DeleteRule(ctx context.Context, rid string) error {
	cond, values, err := builder.BuildDelete(r.table, map[string]interface{}{"id": rid})
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, cond, values...)
	return err
}

func (r *RuleStorage) Export(ctx context.Context) ([]*entity.Rule, error) {
	query, values, _ := builder.BuildSelect(r.table, nil, []string{"*"})
	rows, err := r.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []*entity.Rule
	err = scanner.Scan(rows, &rules)
	return rules, err
}

func (r *RuleStorage) Import(ctx context.Context, rules ...*entity.Rule) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	// 清空数据库
	cond, values, _ := builder.BuildDelete(r.table, nil)
	_, err = tx.ExecContext(ctx, cond, values...)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	var records = make([]map[string]interface{}, len(rules))
	for i, rule := range rules {
		record, err := scanner.Map(rule, "ddb")
		delete(record, "ctime")
		delete(record, "mtime")
		if err != nil {
			_ = tx.Rollback()
			return err
		}
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
