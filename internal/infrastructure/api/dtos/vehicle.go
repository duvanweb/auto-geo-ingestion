package dtos

import "time"

// CreateVehicleRequest is the request body for creating a new vehicle.
type CreateVehicleRequest struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Plate       string `json:"plate"`
}

// UpdateVehicleRequest is the request body for updating an existing vehicle.
type UpdateVehicleRequest struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Plate       string `json:"plate"`
}

// VehicleResponse is the response body for vehicle-related endpoints.
type VehicleResponse struct {
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Plate       string    `json:"plate"`
	UpdatedAt   time.Time `json:"updated_at"`
}
