package deepmock

import (
	"html/template"
	"math/rand"
	"os"
	"time"

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

	// create build-in template functions
	defaultTemplateFuncs = make(template.FuncMap)
	_ = RegisterTemplateFunc("uuid", genUUID)
	_ = RegisterTemplateFunc("timestamp", currentTimestamp)
	_ = RegisterTemplateFunc("date", formatDate)
	_ = RegisterTemplateFunc("plus", plus)
	_ = RegisterTemplateFunc("rand_string", genRandomString)

	// create default hash pool
	defaultHashPoll = newHashPool()
	rand.Seed(time.Now().UnixNano())
}
