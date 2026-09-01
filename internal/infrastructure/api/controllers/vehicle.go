package controllers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"auto-geo-ingestion/internal/core/domain"
	"auto-geo-ingestion/internal/core/ports/services"
	"auto-geo-ingestion/internal/infrastructure/api/dtos"
	apierrors "auto-geo-ingestion/internal/infrastructure/api/errors"
	"auto-geo-ingestion/internal/infrastructure/pkg/logger"
)

// Vehicle is the HTTP controller for vehicle CRUD endpoints.
type Vehicle struct {
	log     logger.Logger
	service services.VehicleService
}

// NewVehicle creates and returns a new Vehicle controller.
func NewVehicle(log logger.Logger, svc services.VehicleService) *Vehicle {
	return &Vehicle{log: log, service: svc}
}

// @Router /v1/vehicles [post]
// @Tags vehicles
// @Summary Create a new vehicle.
// @Param body body dtos.CreateVehicleRequest true "Vehicle creation payload."
// @Success 201 {object} dtos.VehicleResponse "Vehicle created successfully."
// @Failure 400 "Invalid request body."
// @Failure 500 "Unexpected error."
// Create handles POST /v1/vehicles requests and registers a new vehicle.
func (c *Vehicle) Create(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.log.Errorw(r.Context(), "failed to decode create vehicle request", "error", err)
		apierrors.WriteError(w, http.StatusBadRequest, err)
		return
	}

	v, err := c.service.Create(r.Context(), req.Name, req.Plate, req.Description)
	if err != nil {
		apierrors.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if encErr := json.NewEncoder(w).Encode(toVehicleResponse(v)); encErr != nil {
		c.log.Errorw(r.Context(), "failed to encode create vehicle response", "error", encErr)
	}
}

// @Router /v1/vehicles/{id} [delete]
// @Tags vehicles
// @Summary Delete a vehicle by ID.
// @Param id path int true "Vehicle ID."
// @Success 204 "Vehicle deleted successfully."
// @Failure 400 "Invalid vehicle ID."
// @Failure 404 "Vehicle not found."
// @Failure 500 "Unexpected error."
// Delete handles DELETE /v1/vehicles/{id} requests and soft-deletes the vehicle.
func (c *Vehicle) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseVehicleID(r)
	if err != nil {
		c.log.Errorw(r.Context(), "failed to parse vehicle ID for delete", "error", err)
		apierrors.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := c.service.Delete(r.Context(), id); err != nil {
		apierrors.WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Router /v1/vehicles/{id} [get]
// @Tags vehicles
// @Summary Get a vehicle by ID.
// @Param id path int true "Vehicle ID."
// @Success 200 {object} dtos.VehicleResponse "Vehicle found."
// @Failure 400 "Invalid vehicle ID."
// @Failure 404 "Vehicle not found."
// GetByID handles GET /v1/vehicles/{id} requests and returns the matching vehicle.
func (c *Vehicle) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseVehicleID(r)
	if err != nil {
		c.log.Errorw(r.Context(), "failed to parse vehicle ID for get", "error", err)
		apierrors.WriteError(w, http.StatusBadRequest, err)
		return
	}

	v, err := c.service.GetByID(r.Context(), id)
	if err != nil {
		apierrors.WriteError(w, http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if encErr := json.NewEncoder(w).Encode(toVehicleResponse(v)); encErr != nil {
		c.log.Errorw(r.Context(), "failed to encode get vehicle response", "vehicle_id", id, "error", encErr)
	}
}

// @Router /v1/vehicles [get]
// @Tags vehicles
// @Summary List all vehicles.
// @Success 200 {array} dtos.VehicleResponse "List of active vehicles."
// @Failure 500 "Unexpected error."
// List handles GET /v1/vehicles requests and returns all active vehicles.
func (c *Vehicle) List(w http.ResponseWriter, r *http.Request) {
	vehicles, err := c.service.List(r.Context())
	if err != nil {
		apierrors.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	resp := make([]dtos.VehicleResponse, 0, len(vehicles))
	for i := range vehicles {
		resp = append(resp, toVehicleResponse(&vehicles[i]))
	}

	w.Header().Set("Content-Type", "application/json")

	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		c.log.Errorw(r.Context(), "failed to encode list vehicles response", "error", encErr)
	}
}

// @Router /v1/vehicles/{id} [put]
// @Tags vehicles
// @Summary Update a vehicle by ID.
// @Param id path int true "Vehicle ID."
// @Param body body dtos.UpdateVehicleRequest true "Vehicle update payload."
// @Success 200 {object} dtos.VehicleResponse "Vehicle updated successfully."
// @Failure 400 "Invalid request."
// @Failure 404 "Vehicle not found."
// @Failure 500 "Unexpected error."
// Update handles PUT /v1/vehicles/{id} requests and modifies the vehicle's fields.
func (c *Vehicle) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseVehicleID(r)
	if err != nil {
		c.log.Errorw(r.Context(), "failed to parse vehicle ID for update", "error", err)
		apierrors.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var req dtos.UpdateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.log.Errorw(r.Context(), "failed to decode update vehicle request", "vehicle_id", id, "error", err)
		apierrors.WriteError(w, http.StatusBadRequest, err)
		return
	}

	v, err := c.service.Update(r.Context(), id, req.Name, req.Plate, req.Description)
	if err != nil {
		apierrors.WriteError(w, http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if encErr := json.NewEncoder(w).Encode(toVehicleResponse(v)); encErr != nil {
		c.log.Errorw(r.Context(), "failed to encode update vehicle response", "vehicle_id", id, "error", encErr)
	}
}

// parseVehicleID extracts and parses the "id" URL parameter as a uint64.
func parseVehicleID(r *http.Request) (uint64, error) {
	return strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
}

// toVehicleResponse maps a domain.Vehicle to its HTTP response DTO.
func toVehicleResponse(v *domain.Vehicle) dtos.VehicleResponse {
	return dtos.VehicleResponse{
		CreatedAt:   v.CreatedAt,
		Description: v.Description,
		ID:          v.ID,
		Name:        v.Name,
		Plate:       v.Plate,
		UpdatedAt:   v.UpdatedAt,
	}
}
