package testdata

import "auto-geo-ingestion/internal/core/domain"

// GetTestHealth returns a Health fixture for use in tests.
func GetTestHealth() domain.Health {
	return domain.Health{Status: "ok"}
}
