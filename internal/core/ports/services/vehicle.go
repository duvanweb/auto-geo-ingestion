package services

import (
	"context"

	"auto-geo-ingestion/internal/core/domain"
)

//go:generate mockery --name VehicleService --dir=. --output=./mocks

// VehicleService defines the business logic contract for vehicle management.
type VehicleService interface {
	// Create registers a new vehicle with the given attributes.
	Create(ctx context.Context, name, plate, description string) (*domain.Vehicle, error)

	// Delete soft-deletes a vehicle by ID and evicts its cache entries.
	Delete(ctx context.Context, id uint64) error

	// GetByID retrieves a single vehicle by its primary key.
	GetByID(ctx context.Context, id uint64) (*domain.Vehicle, error)

	// List retrieves all active (non-deleted) vehicles.
	List(ctx context.Context) ([]domain.Vehicle, error)

	// Update modifies a vehicle's editable fields and returns the updated record.
	Update(ctx context.Context, id uint64, name, plate, description string) (*domain.Vehicle, error)
}
