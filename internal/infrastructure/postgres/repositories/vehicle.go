package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"auto-geo-ingestion/internal/core/domain"
	portrepos "auto-geo-ingestion/internal/core/ports/repositories"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
)

// ErrVehicleNotFound is returned when no vehicle matches the given criteria.
var ErrVehicleNotFound = errors.New("vehicle not found")

// VehicleRepository implements portrepos.VehicleRepository using PostgreSQL.
type VehicleRepository struct {
	db  Databaser
	log logger.Logger
}

// NewVehicleRepository creates a new VehicleRepository backed by the provided database connection.
func NewVehicleRepository(log logger.Logger, db Databaser) portrepos.VehicleRepository {
	return &VehicleRepository{db: db, log: log}
}

// Create inserts a new vehicle record and returns the persisted entity.
func (r *VehicleRepository) Create(ctx context.Context, v domain.Vehicle) (*domain.Vehicle, error) {
	const query = `
		INSERT INTO vehicles (name, plate, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, plate, description, created_at, updated_at`

	now := time.Now()
	row := r.db.QueryRowContext(ctx, query, v.Name, v.Plate, v.Description, now, now)

	created, err := scanVehicle(row)
	if err != nil {
		r.log.Errorw(ctx, "failed to scan created vehicle", "error", err)
		return nil, err
	}

	return created, nil
}

// Delete soft-deletes a vehicle by setting its deleted_at timestamp.
// Returns ErrVehicleNotFound when no active vehicle with the given ID exists.
func (r *VehicleRepository) Delete(ctx context.Context, id uint64) error {
	const query = `UPDATE vehicles SET deleted_at=$1 WHERE id=$2 AND deleted_at IS NULL`

	res, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		r.log.Errorw(ctx, "failed to soft-delete vehicle", "vehicle_id", id, "error", err)
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		r.log.Errorw(ctx, "failed to read rows affected on delete", "vehicle_id", id, "error", err)
		return err
	}

	if n == 0 {
		return ErrVehicleNotFound
	}

	return nil
}

// FindAll retrieves all non-deleted vehicles ordered by ID.
func (r *VehicleRepository) FindAll(ctx context.Context) ([]domain.Vehicle, error) {
	const query = `
		SELECT id, name, plate, description, created_at, updated_at
		FROM vehicles WHERE deleted_at IS NULL ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.log.Errorw(ctx, "failed to query vehicles", "error", err)
		return nil, err
	}
	defer rows.Close()

	var vehicles []domain.Vehicle

	for rows.Next() {
		var v domain.Vehicle
		if err := rows.Scan(&v.ID, &v.Name, &v.Plate, &v.Description, &v.CreatedAt, &v.UpdatedAt); err != nil {
			r.log.Errorw(ctx, "failed to scan vehicle row", "error", err)
			return nil, err
		}

		vehicles = append(vehicles, v)
	}

	if err := rows.Err(); err != nil {
		r.log.Errorw(ctx, "error iterating vehicle rows", "error", err)
		return nil, err
	}

	return vehicles, nil
}

// FindByID retrieves a single non-deleted vehicle by its primary key.
// Returns ErrVehicleNotFound when no matching active vehicle exists.
func (r *VehicleRepository) FindByID(ctx context.Context, id uint64) (*domain.Vehicle, error) {
	const query = `
		SELECT id, name, plate, description, created_at, updated_at
		FROM vehicles WHERE id=$1 AND deleted_at IS NULL`

	row := r.db.QueryRowContext(ctx, query, id)

	v, err := scanVehicle(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVehicleNotFound
	}

	if err != nil {
		r.log.Errorw(ctx, "failed to scan vehicle by ID", "vehicle_id", id, "error", err)
		return nil, err
	}

	return v, nil
}

// Update modifies a vehicle's editable fields and returns the updated record.
// Returns ErrVehicleNotFound when no active vehicle with the given ID exists.
func (r *VehicleRepository) Update(ctx context.Context, v domain.Vehicle) (*domain.Vehicle, error) {
	const query = `
		UPDATE vehicles SET name=$1, plate=$2, description=$3, updated_at=$4
		WHERE id=$5 AND deleted_at IS NULL
		RETURNING id, name, plate, description, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query, v.Name, v.Plate, v.Description, time.Now(), v.ID)

	updated, err := scanVehicle(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVehicleNotFound
	}

	if err != nil {
		r.log.Errorw(ctx, "failed to scan updated vehicle", "vehicle_id", v.ID, "error", err)
		return nil, err
	}

	return updated, nil
}

// scanVehicle reads a single vehicle row from a *sql.Row.
func scanVehicle(row *sql.Row) (*domain.Vehicle, error) {
	var v domain.Vehicle

	err := row.Scan(&v.ID, &v.Name, &v.Plate, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &v, nil
}
