package domain

import "time"

// Vehicle represents a fleet vehicle in the system.
type Vehicle struct {
	ID          uint64
	Name        string
	Plate       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
