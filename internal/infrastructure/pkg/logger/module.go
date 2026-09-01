package logger

import "go.uber.org/fx"

// Module provides the FX dependency injection module for the logger.
func Module() fx.Option {
	return fx.Module(
		"logger",
		fx.Provide(NewZapLogger),
	)
}
