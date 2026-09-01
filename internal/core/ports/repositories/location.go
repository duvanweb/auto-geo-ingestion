package repositories

import (
	"context"

	"auto-geo-ingestion/internal/core/domain"
)

//go:generate mockery --name LocationRepository --dir=. --output=./mocks

// LocationRepository defines the persistence contract for vehicle location events.
type LocationRepository interface {
	// FindLatestByVehicle retrieves the most recent locations for a vehicle up to the given limit.
	FindLatestByVehicle(ctx context.Context, vehicleID uint64, limit int) ([]domain.Location, error)

	// Save persists a location event and returns the saved record with its generated ID and timestamps.
	Save(ctx context.Context, loc domain.Location) (*domain.Location, error)
}
