package repositories

import (
	"context"

	"auto-geo-ingestion/internal/core/domain"
)

//go:generate mockery --name VehicleRepository --dir=. --output=./mocks

// VehicleRepository defines the persistence contract for vehicle entities.
type VehicleRepository interface {
	// Create persists a new vehicle and returns the created record.
	Create(ctx context.Context, v domain.Vehicle) (*domain.Vehicle, error)

	// FindAll retrieves all non-deleted vehicles ordered by ID.
	FindAll(ctx context.Context) ([]domain.Vehicle, error)

	// FindByID retrieves a single non-deleted vehicle by its primary key.
	FindByID(ctx context.Context, id uint64) (*domain.Vehicle, error)

	// Update modifies an existing vehicle's fields and returns the updated record.
	Update(ctx context.Context, v domain.Vehicle) (*domain.Vehicle, error)

	// Delete soft-deletes a vehicle by setting its deleted_at timestamp.
	Delete(ctx context.Context, id uint64) error
}
