package env

import "github.com/caarlos0/env/v10"

// LoadEnv parses environment variables into the generic config type T.
func LoadEnv[T any]() (*T, error) {
	var config T
	return &config, env.Parse(&config)
}
