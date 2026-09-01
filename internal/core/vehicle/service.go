package vehicle

import (
	"context"

	"auto-geo-ingestion/internal/core/domain"
	"auto-geo-ingestion/internal/core/ports/repositories"
	"auto-geo-ingestion/internal/core/ports/resources"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
)

// Service implements services.VehicleService with PostgreSQL persistence and Redis cache eviction.
type Service struct {
	repo  repositories.VehicleRepository
	cache resources.LocationCache
	log   logger.Logger
}

// NewService creates a new vehicle Service with the given repository, cache, and logger.
func NewService(repo repositories.VehicleRepository, cache resources.LocationCache, log logger.Logger) *Service {
	return &Service{repo: repo, cache: cache, log: log}
}

// Create registers a new vehicle with the given attributes.
func (s *Service) Create(ctx context.Context, name, plate, description string) (*domain.Vehicle, error) {
	v, err := s.repo.Create(ctx, domain.Vehicle{Name: name, Plate: plate, Description: description})
	if err != nil {
		s.log.Errorw(ctx, "failed to create vehicle", "name", name, "plate", plate, "error", err)
		return nil, err
	}

	return v, nil
}

// Delete soft-deletes a vehicle by ID and asynchronously evicts its Redis cache keys.
// The cache eviction is best-effort and does not affect the response to the caller.
func (s *Service) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Errorw(ctx, "failed to delete vehicle", "vehicle_id", id, "error", err)
		return err
	}

	// Best-effort cache eviction: runs in background so the caller is not blocked.
	go func() {
		if err := s.cache.DeleteVehicleKeys(context.Background(), id); err != nil {
			s.log.Errorw(context.Background(), "failed to delete vehicle cache keys", "vehicle_id", id, "error", err)
		}
	}()

	return nil
}

// GetByID retrieves a single vehicle by its primary key.
func (s *Service) GetByID(ctx context.Context, id uint64) (*domain.Vehicle, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.log.Errorw(ctx, "failed to get vehicle by ID", "vehicle_id", id, "error", err)
		return nil, err
	}

	return v, nil
}

// List retrieves all active (non-deleted) vehicles.
func (s *Service) List(ctx context.Context) ([]domain.Vehicle, error) {
	vehicles, err := s.repo.FindAll(ctx)
	if err != nil {
		s.log.Errorw(ctx, "failed to list vehicles", "error", err)
		return nil, err
	}

	return vehicles, nil
}

// Update modifies a vehicle's editable fields and returns the updated record.
func (s *Service) Update(ctx context.Context, id uint64, name, plate, description string) (*domain.Vehicle, error) {
	v, err := s.repo.Update(ctx, domain.Vehicle{ID: id, Name: name, Plate: plate, Description: description})
	if err != nil {
		s.log.Errorw(ctx, "failed to update vehicle", "vehicle_id", id, "error", err)
		return nil, err
	}

	return v, nil
}
