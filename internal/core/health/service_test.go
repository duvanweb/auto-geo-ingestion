package health_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"auto-geo-ingestion/internal/core/domain"
	"auto-geo-ingestion/internal/core/health"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
	testdata "auto-geo-ingestion/test/data"
)

func TestHealthService_GetHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedResult domain.Health
		expectedError  error
	}{
		{
			name:           "works correctly",
			expectedResult: testdata.GetTestHealth(),
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := health.NewService(logger.NewNop())
			result, err := svc.GetHealth(context.Background())

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
