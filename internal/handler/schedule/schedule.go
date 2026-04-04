package schedule

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/schedule"
	"github.com/internships-backend/test-backend-barashF/internal/handler/schedule/validation"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type Controller struct {
	serviceSchedule scheduleService
	logger          logger.Logger
}

func NewController(service scheduleService, logger logger.Logger) *Controller {
	return &Controller{
		serviceSchedule: service,
		logger:          logger,
	}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	var req dto.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	if err := validation.ValidateCreateRequestSchedule(&req); err != nil {
		c.logger.Warn("Create schedule request validation failed",
			logger.F("dayes of week", req.DaysOfWeek),
			logger.F("error", err.Error()),
		)
		c.writeError(w, err, r)
		return
	}

	err = c.serviceSchedule.Create(r.Context(), roomID, model.DaysOfWeek(req.DaysOfWeek), req.StartTime, req.EndTime)
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	c.writeJSON(w, http.StatusCreated, map[string]string{
		"message": "schedule and slots created successfully",
	})
}

func (c *Controller) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		c.logger.Error("Failed to encode response JSON", logger.F("error", err.Error()))
		http.Error(w, `{"error":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}

func (c *Controller) writeError(w http.ResponseWriter, err error, r *http.Request) {
	message, status := c.mapError(err)
	if status == http.StatusInternalServerError {
		c.logger.Error("Internal server error occurred",
			logger.F("error", err.Error()),
			logger.F("path", r.URL.Path),
		)
	} else {
		c.logger.Warn("Request failed with client error",
			logger.F("status", status),
			logger.F("error", err.Error()),
			logger.F("path", r.URL.Path),
		)
	}

	c.writeJSON(w, status, map[string]string{"error": message})
}

func (c *Controller) mapError(err error) (string, int) {
	switch {
	case errors.Is(err, model.ErrMissingRequiredFields),
		errors.Is(err, model.ErrInvalidTime),
		errors.Is(err, model.ErrInvalidDaysOfWeek):
		return err.Error(), http.StatusBadRequest

	case errors.Is(err, model.ErrScheduleExists):
		return err.Error(), http.StatusConflict

	case errors.Is(err, model.ErrNotFound):
		return err.Error(), http.StatusNotFound

	default:
		return "Internal server error", http.StatusInternalServerError
	}
}
