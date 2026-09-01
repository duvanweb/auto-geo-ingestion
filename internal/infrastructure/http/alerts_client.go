package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"auto-geo-ingestion/internal/infrastructure/pkg/env"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"

	jsoniter "github.com/json-iterator/go"
	"github.com/sony/gobreaker"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// AlertsClientConfig holds configuration for the alerts service HTTP client.
type AlertsClientConfig struct {
	AlertsServiceURL string `env:"ALERTS_SERVICE_URL" envDefault:"http://localhost:8081"`
}

// AlertsClient sends vehicle location events to the external alerts service,
// protected by a circuit breaker to prevent cascade failures.
type AlertsClient struct {
	cfg    *AlertsClientConfig
	cb     *gobreaker.CircuitBreaker
	client *http.Client
	log    logger.Logger
}

type detectionPayload struct {
	VehicleID uint64    `json:"vehicle_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
}

// NewAlertsClient constructs an AlertsClient with a circuit breaker configured for the alerts service.
func NewAlertsClient(log logger.Logger) (*AlertsClient, error) {
	cfg, err := env.LoadEnv[AlertsClientConfig]()
	if err != nil {
		return nil, fmt.Errorf("loading alerts client config: %w", err)
	}

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "alerts-service",
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.ConsecutiveFailures > 3
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Errorw(context.Background(), "circuit breaker state change",
				"name", name, "from", from.String(), "to", to.String())
		},
	})

	return &AlertsClient{
		cfg:    cfg,
		cb:     cb,
		client: &http.Client{Timeout: 5 * time.Second},
		log:    log,
	}, nil
}

// NotifyLocation sends a location detection event to the alerts service.
// Failures are logged and silently swallowed — the circuit breaker prevents cascades.
func (c *AlertsClient) NotifyLocation(ctx context.Context, vehicleID uint64, lat, lng float64, timestamp time.Time) {
	payload := detectionPayload{VehicleID: vehicleID, Lat: lat, Lng: lng, Timestamp: timestamp}

	body, err := json.Marshal(payload)
	if err != nil {
		c.log.Errorw(ctx, "failed to marshal detection payload", "vehicle_id", vehicleID, "error", err)
		return
	}

	_, err = c.cb.Execute(func() (any, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			fmt.Sprintf("%s/api/v1/detections", c.cfg.AlertsServiceURL),
			bytes.NewReader(body))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusInternalServerError {
			return nil, fmt.Errorf("alerts service returned %d", resp.StatusCode)
		}

		return nil, nil
	})

	if err != nil {
		c.log.Errorw(ctx, "alerts service notification failed (circuit breaker)",
			"vehicle_id", vehicleID, "error", err)
	}
}
