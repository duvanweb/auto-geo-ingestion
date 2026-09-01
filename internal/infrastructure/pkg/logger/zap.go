package logger

import (
	"context"

	"go.uber.org/zap"
)

type zapAdapter struct {
	log *zap.SugaredLogger
}

func (z *zapAdapter) Debugw(_ context.Context, msg string, keysAndValues ...interface{}) {
	z.log.Debugw(msg, keysAndValues...)
}

func (z *zapAdapter) Errorw(_ context.Context, msg string, keysAndValues ...interface{}) {
	z.log.Errorw(msg, keysAndValues...)
}

func (z *zapAdapter) Fatalw(_ context.Context, msg string, keysAndValues ...interface{}) {
	z.log.Fatalw(msg, keysAndValues...)
}

func (z *zapAdapter) Infow(_ context.Context, msg string, keysAndValues ...interface{}) {
	z.log.Infow(msg, keysAndValues...)
}

func (z *zapAdapter) Warnw(_ context.Context, msg string, keysAndValues ...interface{}) {
	z.log.Warnw(msg, keysAndValues...)
}

// NewZapLogger creates a Logger backed by a zap SugaredLogger.
func NewZapLogger() (Logger, error) {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return &zapAdapter{log: zapLog.Sugar()}, nil
}
