package deepmock

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

const (
	logOutput = "DEEPMOCK_LOGFILE"
)

func init() {
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
