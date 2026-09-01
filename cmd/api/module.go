package main

import (
	"auto-geo-ingestion/internal/core/health"
	"auto-geo-ingestion/internal/infrastructure/api/router"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"

	"go.uber.org/fx"
)

// Module aggregates all FX dependency modules for the application.
func Module() fx.Option {
	return fx.Options(
		logger.Module(),
		health.Module,
		router.Module(),
	)
}
