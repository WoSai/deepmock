package repository

import (
	"database/sql"
	"sync"
	"time"

	"github.com/didi/gendry/manager"
	_ "github.com/go-sql-driver/mysql"
	"github.com/wosai/deepmock"
	"github.com/wosai/deepmock/types"
	"go.uber.org/zap"
)

var (
	db   *sql.DB
	once sync.Once
	// Rule 该模块暴露的全局RuleRepository对象
	Rule types.RuleRepository
)

func BuildDBConnection(opt types.DatabaseOption) *sql.DB {
	once.Do(func() {
		var err error

		for i := 0; i <= opt.ConnectRetry; i++ {
			db, err = manager.New(opt.Name, opt.Username, opt.Password, opt.Host).Set(
				manager.SetCharset("utf8mb4"),
				manager.SetAllowCleartextPasswords(true),
				manager.SetInterpolateParams(true),
				manager.SetTimeout(3*time.Second),
				manager.SetReadTimeout(3*time.Second),
				manager.SetParseTime(true),
				manager.SetLoc("Local"),
			).Port(opt.Port).Open(true)
			if err != nil {
				deepmock.Logger.Error("failed to connect to mysql", zap.Any("params", opt), zap.Error(err))
				time.Sleep(2 * time.Second)
				continue
			}
			deepmock.Logger.Info("accessed mysql database", zap.Any("params", opt))
			break
		}
		if err != nil {
			panic(err)
		}

		// 数据库连接成功
		Rule = NewRuleRepository(db)
	})
	return db
}