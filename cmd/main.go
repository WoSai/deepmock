package main

import (
	"github.com/jacexh/multiconfig"
	"github.com/valyala/fasthttp"
	"github.com/vincentLiuxiang/lu"
	"github.com/wosai/deepmock"
	"go.uber.org/zap"
)

var (
	version = "(git commit revision)"
)

func main() {
	loader := multiconfig.NewWithPathAndEnvPrefix("", "DEEPMOCK")
	opt := new(deepmock.Option)
	loader.MustLoad(opt)

	// 连接数据库
	deepmock.BuildRuleStorage(opt.DB)

	app := lu.New()
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

	deepmock.Logger.Info("deepmock will listen on port "+opt.Server.Port, zap.String("version", version))

	if opt.Server.KeyFile != "" && opt.Server.CertFile != "" {
		deepmock.Logger.Fatal("deepmock is down", zap.Error(
			server.ListenAndServeTLS(opt.Server.Port, opt.Server.CertFile, opt.Server.KeyFile),
		))
	} else {
		deepmock.Logger.Fatal("deepmock is down", zap.Error(server.ListenAndServe(opt.Server.Port)))
	}
}
