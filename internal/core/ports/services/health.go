package services

import (
	"context"

	"auto-geo-ingestion/internal/core/domain"
)

//go:generate mockery --name HealthService --dir=. --output=./mocks

// HealthService defines the contract for health check operations.
type HealthService interface {
	// GetHealth returns the current health status of the service.
	GetHealth(ctx context.Context) (domain.Health, error)
}
