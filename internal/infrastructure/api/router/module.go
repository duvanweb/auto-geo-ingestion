package router

import (
	"auto-geo-ingestion/internal/infrastructure/api/controllers"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

// Module provides the FX dependency injection module for the HTTP API layer.
func Module() fx.Option {
	return fx.Module(
		"api",
		fx.Provide(
			chi.NewRouter,
			NewRouter,
			controllers.NewHealth,
			controllers.NewLocation,
		),
		fx.Invoke(
			registerHooks,
		),
	)
}

func registerHooks(
	lc fx.Lifecycle,
	shutdown fx.Shutdowner,
	router *Router,
) {
	server := &http.Server{
		Addr:    ":8080",
		Handler: router.start("/api"),
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					_ = shutdown.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
