package services

import (
	"context"
	"time"

	"auto-geo-ingestion/internal/core/domain"
)

//go:generate mockery --name LocationIngestionService --dir=. --output=./mocks

// LocationIngestionService defines the business logic contract for ingesting vehicle location data.
type LocationIngestionService interface {
	// Ingest stores a new location event for the given vehicle.
	Ingest(ctx context.Context, vehicleID uint64, lat, lng float64, timestamp time.Time) (*domain.Location, error)
}
