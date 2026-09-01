package dtos

import "time"

// LocationRequest represents the payload for a vehicle location ingestion request.
type LocationRequest struct {
	VehicleID uint64    `json:"vehicle_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
}

// LocationResponse represents the response body for a successfully ingested location.
type LocationResponse struct {
	ID        uint64    `json:"id"`
	VehicleID uint64    `json:"vehicle_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}
