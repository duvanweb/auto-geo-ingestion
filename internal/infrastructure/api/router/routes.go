package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"

	"auto-geo-ingestion/internal/infrastructure/api/controllers"
)

// Controllers holds all HTTP controllers injected via FX.
type Controllers struct {
	fx.In

	Health  *controllers.Health
	Vehicle *controllers.Vehicle
}

// Router wraps the chi mux and its registered controllers.
type Router struct {
	controllers Controllers
	server      *chi.Mux
}

// NewRouter creates a Router with the given mux and controllers.
func NewRouter(server *chi.Mux, c Controllers) *Router {
	return &Router{controllers: c, server: server}
}

// start registers all routes and returns the HTTP handler.
func (r *Router) start(basePath string) http.Handler {
	r.server.Use(middleware.Logger)
	r.server.Use(middleware.Recoverer)

	r.server.Route(basePath, func(route chi.Router) {
		route.Get("/health", r.controllers.Health.GetHealth)

		route.Route("/v1/vehicles", func(vr chi.Router) {
			vr.Post("/", r.controllers.Vehicle.Create)
			vr.Get("/", r.controllers.Vehicle.List)
			vr.Get("/{id}", r.controllers.Vehicle.GetByID)
			vr.Put("/{id}", r.controllers.Vehicle.Update)
			vr.Delete("/{id}", r.controllers.Vehicle.Delete)
		})
	})

	return r.server
}
