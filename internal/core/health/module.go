package health

import (
	"go.uber.org/fx"

	"auto-geo-ingestion/internal/core/ports/services"
)

// Module provides the FX dependency injection module for the health domain.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewService,
			fx.As(new(services.HealthService)),
		),
	),
)
