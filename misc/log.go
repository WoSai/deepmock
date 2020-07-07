package misc

import (
	"os"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
	// Logger DeepMock全局日志对象
	Logger *zap.Logger

	logOutput = "DEEPMOCK_LOGFILE"
)

func init() {
	// create Logger
	conf := zap.NewProductionConfig()
	conf.Sampling = nil
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	conf.EncoderConfig.TimeKey = "@timestamp"

	outputFromEnv := os.Getenv(logOutput)
	if !(outputFromEnv == "") {
		conf.OutputPaths = []string{outputFromEnv}
		conf.ErrorOutputPaths = []string{outputFromEnv}
	}

	var err error
	Logger, err = conf.Build()
	if err != nil {
		panic(err)
	}

	Logger.Info("deepmock logger is initialized")
}
