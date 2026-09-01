package redis

import (
	portresources "auto-geo-ingestion/internal/core/ports/resources"
	"auto-geo-ingestion/internal/infrastructure/pkg/env"
	redisresources "auto-geo-ingestion/internal/infrastructure/redis/resources"

	"go.uber.org/fx"
)

// Configuration holds the Redis connection configuration.
type Configuration struct {
	Host     string `env:"REDIS_HOST" envDefault:"localhost"`
	Port     string `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

// Module returns the FX module for the Redis infrastructure.
func Module() fx.Option {
	return fx.Module(
		"redis",
		fx.Provide(
			env.LoadEnv[Configuration],
			NewClient,
			fx.Annotate(
				redisresources.NewLocationCache,
				fx.As(new(portresources.LocationCache)),
			),
		),
	)
}
