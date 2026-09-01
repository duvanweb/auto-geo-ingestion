package logger

import "context"

// Logger defines the application logging interface using structured key-value pairs.
// Every method accepts a context.Context as the first argument to allow
// propagation of trace IDs or other context values in the future.
type Logger interface {
	Debugw(ctx context.Context, msg string, keysAndValues ...interface{})
	Errorw(ctx context.Context, msg string, keysAndValues ...interface{})
	Fatalw(ctx context.Context, msg string, keysAndValues ...interface{})
	Infow(ctx context.Context, msg string, keysAndValues ...interface{})
	Warnw(ctx context.Context, msg string, keysAndValues ...interface{})
}
