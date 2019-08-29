package main

import (
	"flag"

	"github.com/qastub/deepmock"
	"github.com/vincentLiuxiang/lu"
	"go.uber.org/zap"
)

var (
	onPort     = ":16600"
	datasource = "localhost:3306"
)

func init() {
	flag.StringVar(&onPort, "port", onPort, "deepmock监听端口")
	flag.StringVar(&datasource, "datasource", datasource, "数据库连接地址")
}

func main() {
	flag.Parse()

	app := lu.New()

	app.Get("/api/v1/rule", deepmock.HandleGetRule)
	app.Post("/api/v1/rule", deepmock.HandleCreateRule)
	app.Put("/api/v1/rule", deepmock.HandleUpdateRule)
	app.Patch("/app/v1/rule", deepmock.HandlePatchRule)
	app.Delete("/app/v1/rule", deepmock.HandleDeleteRule)

	app.Get("/app/v1/rules", deepmock.HandleExportRules)
	app.Post("/app/v1/rules", deepmock.HandleImportRules)

	app.Use("/", deepmock.HandleMockedAPI)

	deepmock.Logger.Info("deepmock will listen on port " + onPort)
	deepmock.Logger.Fatal("deepmock is down", zap.Error(app.Listen(onPort)))
}
