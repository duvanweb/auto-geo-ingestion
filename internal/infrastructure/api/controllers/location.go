package controllers

import (
	"errors"
	"net/http"

	"auto-geo-ingestion/internal/core/location"
	"auto-geo-ingestion/internal/core/ports/services"
	apierrors "auto-geo-ingestion/internal/infrastructure/api/errors"
	"auto-geo-ingestion/internal/infrastructure/api/dtos"
	infrahttp "auto-geo-ingestion/internal/infrastructure/http"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
)

// Location handles HTTP requests for vehicle location ingestion.
type Location struct {
	log          logger.Logger
	service      services.LocationIngestionService
	alertsClient *infrahttp.AlertsClient
}

// NewLocation creates a Location controller with the provided dependencies.
func NewLocation(log logger.Logger, svc services.LocationIngestionService, alertsClient *infrahttp.AlertsClient) *Location {
	return &Location{log: log, service: svc, alertsClient: alertsClient}
}

// Ingest processes a vehicle location event, applies deduplication, and stores the result.
// Returns 409 Conflict on duplicates, 202 Accepted when the location is queued via fallback,
// and 201 Created on successful persistence.
func (c *Location) Ingest(w http.ResponseWriter, r *http.Request) {
	var req dtos.LocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.WriteError(w, http.StatusBadRequest, err)
		return
	}

	saved, err := c.service.Ingest(r.Context(), req.VehicleID, req.Lat, req.Lng, req.Timestamp)
	if err != nil {
		if errors.Is(err, location.ErrDuplicate) {
			apierrors.WriteError(w, http.StatusConflict, err)
			return
		}

		c.log.Errorw(r.Context(), "failed to ingest location", "vehicle_id", req.VehicleID, "error", err)
		apierrors.WriteError(w, http.StatusInternalServerError, err)

		return
	}

	// Notify the alerts service asynchronously; the circuit breaker handles failures internally.
	// context.WithoutCancel detaches from the request lifetime so the goroutine is not
	// cancelled the moment the HTTP response is flushed.
	notifyCtx := r.Context()
	go c.alertsClient.NotifyLocation(notifyCtx, req.VehicleID, req.Lat, req.Lng, req.Timestamp)

	if saved == nil {
		// The CB fallback queue accepted the location; acknowledge with 202.
		w.WriteHeader(http.StatusAccepted)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(dtos.LocationResponse{
		ID:        saved.ID,
		VehicleID: saved.VehicleID,
		Lat:       saved.Lat,
		Lng:       saved.Lng,
		Timestamp: saved.Timestamp,
		CreatedAt: saved.CreatedAt,
	})
}
