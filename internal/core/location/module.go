package location

import (
	"auto-geo-ingestion/internal/core/ports/services"

	"go.uber.org/fx"
)

// Module wires the location ingestion Service into the FX dependency graph.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewService, fx.As(new(services.LocationIngestionService))),
	),
)
