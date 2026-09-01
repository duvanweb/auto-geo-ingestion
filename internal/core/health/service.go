package health

import (
	"context"

	"auto-geo-ingestion/internal/core/domain"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
)

// Service implements the HealthService port.
type Service struct {
	logger logger.Logger
}

// GetHealth returns the current health status of the service.
func (s *Service) GetHealth(_ context.Context) (domain.Health, error) {
	return domain.Health{Status: "ok"}, nil
}

// NewService creates and returns a new health Service.
func NewService(log logger.Logger) *Service {
	return &Service{logger: log}
}
