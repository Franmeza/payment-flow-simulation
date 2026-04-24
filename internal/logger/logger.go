package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(service string) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	base, err := config.Build()
	if err != nil {
		panic(err)
	}

	// Attach service name to every log line
	Log = base.With(zap.String("service", service))
}