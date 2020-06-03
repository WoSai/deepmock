package main

import (
	"github.com/jacexh/multiconfig"
	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock"
	"github.com/wosai/deepmock/repository"
	"github.com/wosai/deepmock/router"
	"github.com/wosai/deepmock/types"
	"go.uber.org/zap"
)

var (
	version = "(git commit revision)"
)

func main() {
	loader := multiconfig.NewWithPathAndEnvPrefix("", "DEEPMOCK")
	opt := new(types.Option)
	loader.MustLoad(opt)

	// 连接数据库
	repository.BuildDBConnection(opt.DB)

	// 初始化http handler
	app := router.BuildRouter()
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
