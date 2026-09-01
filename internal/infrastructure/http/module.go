package http

import (
	"go.uber.org/fx"
)

// Module provides the FX dependency injection module for outbound HTTP clients.
func Module() fx.Option {
	return fx.Module("http-clients",
		fx.Provide(NewAlertsClient),
	)
}
