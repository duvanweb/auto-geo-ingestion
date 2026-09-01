package domain

import "time"

// Location represents the geographic position of a vehicle at a given point in time.
type Location struct {
	ID        uint64
	VehicleID uint64
	Lat       float64
	Lng       float64
	Timestamp time.Time
	CreatedAt time.Time
}
