package main

import (
	"flag"

	"github.com/valyala/fasthttp"

	"github.com/qastub/deepmock"
	"github.com/vincentLiuxiang/lu"
	"go.uber.org/zap"
)

var (
	onPort     = ":16600"
	datasource = "localhost:3306"
	version    = "(git commit revision)"
)

func init() {
	flag.StringVar(&onPort, "port", onPort, "监听端口")
	flag.StringVar(&datasource, "datasource", datasource, "数据库连接地址")
}

func main() {
	flag.Parse()

	app := lu.New()

	app.Get("/api/help", deepmock.HandleHelp)

	app.Get("/api/v1/rule", deepmock.HandleGetRule)
	app.Post("/api/v1/rule", deepmock.HandleCreateRule)
	app.Put("/api/v1/rule", deepmock.HandleUpdateRule)
	app.Patch("/api/v1/rule", deepmock.HandlePatchRule)
	app.Delete("/api/v1/rule", deepmock.HandleDeleteRule)

	app.Get("/api/v1/rules", deepmock.HandleExportRules)
	app.Post("/api/v1/rules", deepmock.HandleImportRules)

	app.Use("/", deepmock.HandleMockedAPI)

	server := &fasthttp.Server{
		Name:        "DeepMock Service",
		Handler:     app.Handler,
		Concurrency: 1024 * 1024,
	}

	deepmock.Logger.Info("deepmock will listen on port "+onPort, zap.String("version", version))
	deepmock.Logger.Fatal("deepmock is down", zap.Error(server.ListenAndServe(onPort)))
}
