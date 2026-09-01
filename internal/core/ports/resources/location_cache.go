package resources

//go:generate mockery --name LocationCache --dir=. --output=./mocks

import (
	"context"
	"time"

	"auto-geo-ingestion/internal/core/domain"
)

// LocationCache defines the contract for caching vehicle location data.
type LocationCache interface {
	// DeleteVehicleKeys removes all cached keys associated with a vehicle.
	DeleteVehicleKeys(ctx context.Context, vehicleID uint64) error

	// GetLastLocation retrieves the last known location for a vehicle.
	// Returns false as the second value when no entry exists in the cache.
	GetLastLocation(ctx context.Context, vehicleID uint64) (domain.Location, bool, error)

	// SetLastLocation stores the last known location for a vehicle with the given TTL.
	SetLastLocation(ctx context.Context, loc domain.Location, ttl time.Duration) error
}
