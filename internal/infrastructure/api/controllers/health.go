package controllers

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"

	"auto-geo-ingestion/internal/core/ports/services"
	"auto-geo-ingestion/internal/infrastructure/api/dtos"
	apierrors "auto-geo-ingestion/internal/infrastructure/api/errors"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Health is the HTTP controller for health-related endpoints.
type Health struct {
	logger  logger.Logger
	service services.HealthService
}

// @Router /health [get]
// @Tags health
// @Summary Get service health status.
// @Success 200 {object} dtos.HealthResponse "Service is healthy."
// @Failure 500 "Unexpected error."
// GetHealth handles GET /health requests and returns the current health status.
func (c *Health) GetHealth(w http.ResponseWriter, r *http.Request) {
	health, err := c.service.GetHealth(r.Context())
	if err != nil {
		c.logger.Errorw(r.Context(), "failed to get health status", "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := dtos.HealthResponse{Status: health.Status}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		c.logger.Errorw(r.Context(), "failed to encode health response", "error", encErr)
	}
}

// NewHealth creates and returns a new Health controller.
func NewHealth(log logger.Logger, svc services.HealthService) *Health {
	return &Health{
		logger:  log,
		service: svc,
	}
}
