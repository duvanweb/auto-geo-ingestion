package logger

import "context"

type noopLogger struct{}

func (n *noopLogger) Debugw(_ context.Context, _ string, _ ...interface{}) {}
func (n *noopLogger) Errorw(_ context.Context, _ string, _ ...interface{}) {}
func (n *noopLogger) Fatalw(_ context.Context, _ string, _ ...interface{}) {}
func (n *noopLogger) Infow(_ context.Context, _ string, _ ...interface{})  {}
func (n *noopLogger) Warnw(_ context.Context, _ string, _ ...interface{})  {}

// NewNop returns a Logger that silently discards all output. Intended for use in tests.
func NewNop() Logger {
	return &noopLogger{}
}
