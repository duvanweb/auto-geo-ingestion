package postgres

import (
	portrepos "auto-geo-ingestion/internal/core/ports/repositories"
	"auto-geo-ingestion/internal/infrastructure/pkg/env"
	pgrepositories "auto-geo-ingestion/internal/infrastructure/postgres/repositories"

	"go.uber.org/fx"
)

// Configuration holds PostgreSQL connection configuration loaded from environment variables.
type Configuration struct {
	DBHost string `env:"DB_HOST" envDefault:"localhost"`
	DBPort string `env:"DB_PORT" envDefault:"5432"`
	DBPass string `env:"DB_PASS"`
	DBName string `env:"DB_NAME" envDefault:"auto_geo_ingestion"`
	DBUser string `env:"DB_USER" envDefault:"postgres"`
}

// Module provides the FX dependency injection module for the PostgreSQL infrastructure.
func Module() fx.Option {
	return fx.Module(
		"postgres",
		fx.Provide(
			env.LoadEnv[Configuration],
			NewConnection,
			fx.Annotate(pgrepositories.NewVehicleRepository, fx.As(new(portrepos.VehicleRepository))),
			pgrepositories.NewLocationRepository,
			fx.Annotate(pgrepositories.NewLocationRepositoryWithCB, fx.As(new(portrepos.LocationRepository))),
		),
	)
}
