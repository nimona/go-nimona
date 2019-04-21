package log

import (
	"context"
	"os"
	"strings"

	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
)

var DefaultLogger *zap.Logger

func init() {
	config := zap.NewDevelopmentConfig()
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "INFO":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "WARN":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	DefaultLogger, _ = config.Build()
}

// Logger returns a zap logger with as much context as possible
func Logger(ctx context.Context) *zap.Logger {
	nl := DefaultLogger
	if ctx == nil {
		return nl
	}
	if nctx, ok := ctx.(interface{ CorrelationID() string }); ok {
		nl = nl.With(zap.String("ctx.correlation_id", nctx.CorrelationID()))
	}
	return nl
}
