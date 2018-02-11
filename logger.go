package fabric

import (
	"context"

	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
)

var baseLogger *zap.Logger

func init() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	baseLogger, _ = config.Build()
}

// Logger returns a zap logger with as much context as possible
func Logger(ctx context.Context) *zap.Logger {
	nl := baseLogger
	if ctx != nil {
		if rid, ok := ctx.Value(RequestIDKey{}).(string); ok {
			nl = nl.With(zap.String("req.id", rid))
		}
	}
	return nl
}
