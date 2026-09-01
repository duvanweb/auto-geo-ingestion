package postgres

import (
	"auto-geo-ingestion/internal/infrastructure/pkg/env"

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
		),
	)
}
