package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jacexh/multiconfig"
	"github.com/valyala/fasthttp"
	"github.com/wosai/deepmock/application"
	"github.com/wosai/deepmock/infrastructure"
	"github.com/wosai/deepmock/misc"
	"github.com/wosai/deepmock/option"
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
	db := infrastructure.BuildDBConnection(opt.DB)

	// 初始化service
	application.BuildMockApplication(
		infrastructure.NewRuleRepository(db),
		nil,
		nil,
	)

	// 初始化http handler
	app := router.BuildRouter()
	server := &fasthttp.Server{
		Name:        "DeepMock Service",
		Handler:     app.Handler,
		Concurrency: 1024 * 1024,
	}
	misc.Logger.Info("deepmock is running on port "+opt.Server.Port, zap.String("version", version))

	errChan := make(chan error, 1)
	go func() {
		if opt.Server.KeyFile != "" && opt.Server.CertFile != "" {
			errChan <- server.ListenAndServeTLS(opt.Server.Port, opt.Server.CertFile, opt.Server.KeyFile)
		} else {
			errChan <- server.ListenAndServe(opt.Server.Port)
		}

	}()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		errChan <- fmt.Errorf("caught signal: %s", (<-sigs).String())
	}()

	misc.Logger.Panic("deepmock is shutdown", zap.Error(<-errChan))
}
