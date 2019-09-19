package deepmock

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/manager"
	"github.com/didi/gendry/scanner"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qastub/deepmock/types"
	"go.uber.org/zap"
)

var (
	storage *ruleStorage
)

type (
	ruleStorage struct {
		db           *sql.DB
		table        string
		option       connOption
		connectRetry int
		once         sync.Once
	}

	connOption struct {
		Username string
		Password string
		Host     string
		Port     int
		Database string
	}
)

const (
	envDBName     = "DEEPMOCK_DB_NAME"
	envDBUser     = "DEEPMOCK_DB_USER"
	envDBPassword = "DEEPMOCK_DB_PASSWORD"
	envDBHost     = "DEEPMOCK_DB_HOST"
	envDBPort     = "DEEPMOCK_DB_PORT"
)

// EnableRulePersistence 启用规则持久化
func parseConnOption() connOption {
	op := connOption{
		Username: "root",
		Password: "123",
		Host:     "localhost",
		Port:     3306,
		Database: "deepmock",
	}

	if d := os.Getenv(envDBName); d != "" {
		op.Database = d
	}

	if u := os.Getenv(envDBUser); u != "" {
		op.Username = u
	}

	if p := os.Getenv(envDBPassword); p != "" {
		op.Password = p
	}

	if h := os.Getenv(envDBHost); h != "" {
		op.Host = h
	}

	if port := os.Getenv(envDBPort); port != "" {
		po, err := strconv.Atoi(port)
		if err != nil {
			Logger.Error("bad mysql port", zap.String("port", port))
			panic(err)
		}
		op.Port = po
	}
	return op
}

func (rs *ruleStorage) buildConnection() *sql.DB {
	rs.once.Do(func() {
		var db *sql.DB
		var err error
		co := parseConnOption()

		for i := 0; i <= rs.connectRetry; i++ {
			db, err = manager.New(co.Database, co.Username, co.Password, co.Host).Set(
				manager.SetCharset("utf8mb4"),
				manager.SetAllowCleartextPasswords(true),
				manager.SetInterpolateParams(true),
				manager.SetTimeout(3*time.Second),
				manager.SetReadTimeout(3*time.Second),
				manager.SetParseTime(true),
				manager.SetLoc("Local"),
			).Port(co.Port).Open(true)
			if err != nil {
				Logger.Error("failed to connect to mysql", zap.Any("params", co), zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}
			Logger.Info("accessed mysql database", zap.Any("params", co))
			break
		}
		if err != nil {
			panic(err)
		}
		rs.db = db
		rs.option = co
	})
	return rs.db
}

func (rs *ruleStorage) createRule(res *types.ResourceRule) error {
	err := res.Check()
	if err != nil {
		return err
	}

	method := strings.ToUpper(res.Method)
	res.ID = genID([]byte(res.Path), []byte(method))
	res.Method = method

	v, w, r, err := marshalPropertyOfRule(res)
	if err != nil {
		return err
	}

	query, values, err := builder.BuildInsert("rule", []map[string]interface{}{{
		"id":        genID([]byte(res.Path), []byte(method)),
		"path":      res.Path,
		"method":    strings.ToUpper(res.Method),
		"variable":  v,
		"weight":    w,
		"responses": r,
		"version":   0,
	}})
	Logger.Info(query, zap.Any("values", values))
	if err != nil {
		Logger.Error("failed to build sql statement", zap.Error(err))
		return err
	}
	Logger.Info(query, zap.Any("values", values))
	_, err = rs.buildConnection().Exec(query, values...)
	return err
}

func (rs *ruleStorage) getRule(ri string) (*types.ResourceRule, error) {
	query, values, _ := builder.BuildSelect(
		rs.table,
		map[string]interface{}{"id": ri, "_limit": []uint{1}},
		[]string{"*"},
	)
	Logger.Info(query, zap.Any("values", values))
	rows, err := rs.buildConnection().Query(query, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []*types.ResourceRule
	err = scanner.Scan(rows, &rules)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, errors.New("cannot find rule by id: " + ri)
	}
	return rules[0], nil
}

func (rs *ruleStorage) put(res *types.ResourceRule) (*types.ResourceRule, error) {
	rule, err := rs.getRule(res.ID)
	if err != nil {
		return nil, err
	}

	res.ID = rule.ID
	res.Path = rule.Path
	res.Method = rule.Method
	res.Version = rule.Version
	if err = res.Check(); err != nil {
		return nil, err
	}

	err = rs.updateRecord(res)
	return res, err
}

func (rs *ruleStorage) patch(res *types.ResourceRule) (*types.ResourceRule, error) {
	rule, err := rs.getRule(res.ID)
	if err != nil {
		return nil, err
	}
	if res.Variable != nil {
		m := make(types.ResourceVariable)
		if rule.Variable != nil {
			m = *rule.Variable
		}
		for k, v := range *res.Variable {
			m[k] = v
		}
		rule.Variable = &m
	}

	if res.Weight != nil {
		m := make(types.ResourceWeight)
		if rule.Weight != nil {
			m = *rule.Weight
		}
		for k, v := range *res.Weight {
			d, exist := m[k]
			if !exist {
				m[k] = v
			} else {
				for i, j := range v {
					d[i] = j
				}
			}
		}
	}

	if res.Responses != nil {
		*rule.Responses = *res.Responses
	}

	if err = rule.Check(); err != nil {
		return nil, err
	}

	err = rs.updateRecord(rule)
	return rule, err
}

func (rs *ruleStorage) delete(res *types.ResourceRule) error {
	cond, values, err := builder.BuildDelete(rs.table, map[string]interface{}{"id": res.ID})
	if err != nil {
		Logger.Error("failed to build delete statement", zap.Error(err))
		return err
	}
	Logger.Info(cond, zap.Any("values", values))
	_, err = rs.buildConnection().Exec(cond, values...)
	return err
}

func (rs *ruleStorage) export() ([]*types.ResourceRule, error) {
	query, values, _ := builder.BuildSelect(rs.table, nil, []string{"*"})
	Logger.Info(query, zap.Any("values", values))

	rows, err := rs.buildConnection().Query(query, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []*types.ResourceRule
	err = scanner.Scan(rows, &rules)
	return rules, err
}

func (rs *ruleStorage) importRules(rules ...*types.ResourceRule) error {
	var err error
	for _, rule := range rules {
		if err = rule.Check(); err != nil {
			return err
		}
	}

	db := rs.buildConnection()
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		Logger.Error("failed to create transaction", zap.Error(err))
		return err
	}

	// 清空数据库
	crond, values, _ := builder.BuildDelete(rs.table, nil)
	_, err = tx.Exec(crond, values...)
	if err != nil {
		Logger.Error("failed to delete all rules", zap.Error(err))
		_ = tx.Rollback()
		return err
	}
	var records = make([]map[string]interface{}, len(rules))
	for i, rule := range rules {
		method := strings.ToUpper(rule.Method)
		rule.ID = genID([]byte(rule.Path), []byte(method))
		rule.Method = method

		v, w, r, err := marshalPropertyOfRule(rule)
		if err != nil {
			return err
		}
		records[i] = map[string]interface{}{
			"id":        rule.ID,
			"path":      rule.Path,
			"method":    rule.Method,
			"variable":  v,
			"weight":    w,
			"responses": r,
			"version":   0,
		}
	}

	crond, values, err = builder.BuildInsert(rs.table, records)
	if err != nil {
		Logger.Error("failed to build batch insert sql statement", zap.Error(err))
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec(crond, values...)
	if err != nil {
		Logger.Error("failed to insert batch data", zap.Error(err))
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (rs *ruleStorage) updateRecord(res *types.ResourceRule) error {
	v, w, r, err := marshalPropertyOfRule(res)
	if err != nil {
		return err
	}

	cond, values, err := builder.BuildUpdate(
		rs.table,
		map[string]interface{}{"id": res.ID},
		map[string]interface{}{
			"variable":  v,
			"weight":    w,
			"responses": r,
			"version":   res.Version + 1,
		},
	)
	if err != nil {
		return err
	}
	Logger.Info(cond, zap.Any("values", values))
	_, err = rs.buildConnection().Exec(cond, values...)
	if err != nil {
		return err
	}
	return nil
}

func marshalPropertyOfRule(res *types.ResourceRule) ([]byte, []byte, []byte, error) {
	var v, w, r []byte
	var err error
	if res.Variable != nil {
		if v, err = json.Marshal(res.Variable); err != nil {
			return v, w, r, err
		}
	}
	if res.Weight != nil {
		if w, err = json.Marshal(res.Weight); err != nil {
			return v, w, r, err
		}
	}
	if res.Responses != nil {
		if r, err = json.Marshal(res.Responses); err != nil {
			return v, w, r, err
		}
	}
	return v, w, r, nil
}

func init() {
	storage = &ruleStorage{
		table:        "rule",
		connectRetry: 3,
	}
}