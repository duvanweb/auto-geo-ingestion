package location

import (
	"context"
	"errors"
	"time"

	"auto-geo-ingestion/internal/core/domain"
	"auto-geo-ingestion/internal/core/ports/repositories"
	"auto-geo-ingestion/internal/core/ports/resources"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
)

// ErrDuplicate is returned when an incoming location has the same coordinates as the last known
// position within the deduplication window.
var ErrDuplicate = errors.New("duplicate location: same coordinates within dedup window")

// Service implements the LocationIngestionService business logic.
type Service struct {
	repo     repositories.LocationRepository
	cache    resources.LocationCache
	log      logger.Logger
	dedupTTL time.Duration
}

// NewService creates a new location ingestion Service with a 30-second deduplication TTL.
func NewService(repo repositories.LocationRepository, cache resources.LocationCache, log logger.Logger) *Service {
	return &Service{
		repo:     repo,
		cache:    cache,
		log:      log,
		dedupTTL: 30 * time.Second,
	}
}

// Ingest stores a new location event for the given vehicle, applying anti-duplicate checks.
// It returns ErrDuplicate when the incoming coordinates match the last cached position.
func (s *Service) Ingest(ctx context.Context, vehicleID uint64, lat, lng float64, timestamp time.Time) (*domain.Location, error) {
	last, found, err := s.cache.GetLastLocation(ctx, vehicleID)
	if err != nil {
		s.log.Errorw(ctx, "failed to get last location from cache", "vehicle_id", vehicleID, "error", err)
	}

	if found && isSamePosition(last.Lat, last.Lng, lat, lng) {
		return nil, ErrDuplicate
	}

	loc := domain.Location{
		VehicleID: vehicleID,
		Lat:       lat,
		Lng:       lng,
		Timestamp: timestamp,
	}

	saved, err := s.repo.Save(ctx, loc)
	if err != nil {
		s.log.Errorw(ctx, "failed to save location", "vehicle_id", vehicleID, "error", err)
		return nil, err
	}

	if err := s.cache.SetLastLocation(ctx, *saved, s.dedupTTL); err != nil {
		s.log.Errorw(ctx, "failed to update location cache", "vehicle_id", vehicleID, "error", err)
	}

	return saved, nil
}

// isSamePosition reports whether two geographic positions are within the deduplication epsilon.
func isSamePosition(lat1, lng1, lat2, lng2 float64) bool {
	const epsilon = 0.00001
	return absoluteDiff(lat1, lat2) < epsilon && absoluteDiff(lng1, lng2) < epsilon
}

// absoluteDiff returns the absolute difference between two float64 values.
func absoluteDiff(a, b float64) float64 {
	diff := a - b
	if diff < 0 {
		return -diff
	}

	return diff
}
