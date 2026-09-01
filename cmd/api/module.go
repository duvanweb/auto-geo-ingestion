package main

import (
	"auto-geo-ingestion/internal/core/health"
	"auto-geo-ingestion/internal/core/location"
	"auto-geo-ingestion/internal/core/vehicle"
	"auto-geo-ingestion/internal/infrastructure/api/router"
	infrahttp "auto-geo-ingestion/internal/infrastructure/http"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
	"auto-geo-ingestion/internal/infrastructure/postgres"
	infraredis "auto-geo-ingestion/internal/infrastructure/redis"

	"go.uber.org/fx"
)

// Module aggregates all FX dependency modules for the application.
func Module() fx.Option {
	return fx.Options(
		logger.Module(),
		postgres.Module(),
		infraredis.Module(),
		infrahttp.Module(),
		health.Module,
		location.Module,
		vehicle.Module,
		router.Module(),
	)
}
