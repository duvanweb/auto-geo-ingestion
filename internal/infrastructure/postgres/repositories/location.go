package repositories

import (
	"context"

	"auto-geo-ingestion/internal/core/domain"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
)

// LocationRepository implements the location persistence contract using PostgreSQL.
type LocationRepository struct {
	db  Databaser
	log logger.Logger
}

// NewLocationRepository creates a new LocationRepository backed by the provided database connection.
func NewLocationRepository(log logger.Logger, db Databaser) *LocationRepository {
	return &LocationRepository{db: db, log: log}
}

// FindLatestByVehicle retrieves the most recent locations for a vehicle up to the given limit.
func (r *LocationRepository) FindLatestByVehicle(ctx context.Context, vehicleID uint64, limit int) ([]domain.Location, error) {
	const query = `
		SELECT id, vehicle_id, lat, lng, timestamp, created_at
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY timestamp DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, vehicleID, limit)
	if err != nil {
		r.log.Errorw(ctx, "failed to query latest locations", "vehicle_id", vehicleID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var locs []domain.Location

	for rows.Next() {
		var l domain.Location
		if err := rows.Scan(&l.ID, &l.VehicleID, &l.Lat, &l.Lng, &l.Timestamp, &l.CreatedAt); err != nil {
			r.log.Errorw(ctx, "failed to scan location row", "vehicle_id", vehicleID, "error", err)
			return nil, err
		}

		locs = append(locs, l)
	}

	if err := rows.Err(); err != nil {
		r.log.Errorw(ctx, "error iterating location rows", "vehicle_id", vehicleID, "error", err)
		return nil, err
	}

	return locs, nil
}

// Save persists a location event and returns the saved record with its generated ID and timestamps.
func (r *LocationRepository) Save(ctx context.Context, loc domain.Location) (*domain.Location, error) {
	const query = `
		INSERT INTO vehicle_locations (vehicle_id, lat, lng, timestamp, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, vehicle_id, lat, lng, timestamp, created_at`

	row := r.db.QueryRowContext(ctx, query, loc.VehicleID, loc.Lat, loc.Lng, loc.Timestamp)

	var saved domain.Location
	if err := row.Scan(&saved.ID, &saved.VehicleID, &saved.Lat, &saved.Lng, &saved.Timestamp, &saved.CreatedAt); err != nil {
		r.log.Errorw(ctx, "failed to scan saved location", "vehicle_id", loc.VehicleID, "error", err)
		return nil, err
	}

	return &saved, nil
}
