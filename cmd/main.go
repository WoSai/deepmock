package main

import (
	"github.com/jacexh/multiconfig"
	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock/application"
	"github.com/wosai/deepmock/misc"
	"github.com/wosai/deepmock/option"
	"github.com/wosai/deepmock/repository"
	"github.com/wosai/deepmock/router"
	"go.uber.org/zap"
)

var (
	version = "(git commit revision)"
)

func main() {
	loader := multiconfig.NewWithPathAndEnvPrefix("", "DEEPMOCK")
	opt := new(option.Option)
	loader.MustLoad(opt)

	// 连接数据库
	repository.BuildDBConnection(opt.DB)

	// 初始化service
	application.BuildRuleService(repository.Rule)

	// 初始化http handler
	app := router.BuildRouter()
	server := &fasthttp.Server{
		Name:        "DeepMock Service",
		Handler:     app.Handler,
		Concurrency: 1024 * 1024,
	}
	misc.Logger.Info("deepmock will listen on port "+opt.Server.Port, zap.String("version", version))

	if opt.Server.KeyFile != "" && opt.Server.CertFile != "" {
		misc.Logger.Fatal("deepmock is down", zap.Error(
			server.ListenAndServeTLS(opt.Server.Port, opt.Server.CertFile, opt.Server.KeyFile),
		))
	} else {
		misc.Logger.Fatal("deepmock is down", zap.Error(server.ListenAndServe(opt.Server.Port)))
	}
}
