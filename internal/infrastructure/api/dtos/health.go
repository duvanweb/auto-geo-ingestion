package dtos

// HealthResponse is the response body for health check endpoints.
type HealthResponse struct {
	Status string `json:"status"`
}
