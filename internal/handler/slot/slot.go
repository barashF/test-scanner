package slot

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/slot"
	"github.com/internships-backend/test-backend-barashF/internal/handler/slot/validation"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type Controller struct {
	serviceSlot slotService
	logger      logger.Logger
}

func NewController(service slotService, logger logger.Logger) *Controller {
	return &Controller{
		serviceSlot: service,
		logger:      logger,
	}
}

func (c *Controller) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid room ID format",
		})
		return
	}

	dateStr := r.URL.Query().Get("date")
	targetDate, err := validation.ValidateSlotDate(dateStr)
	if err != nil {
		c.logger.Warn("Failed to parse date from query",
			logger.F("room_id", roomID),
			logger.F("date_str", dateStr),
			logger.F("error", err.Error()),
		)
		c.writeError(w, err, r)
		return
	}

	slots, err := c.serviceSlot.GetAvailableSlotsByDate(r.Context(), roomID, targetDate)
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	slotsResponse := make([]dto.SlotResponse, 0, len(slots))
	for _, slot := range slots {
		slotsResponse = append(slotsResponse, dto.ModelToResponse(slot))
	}

	c.writeJSON(w, http.StatusOK, map[string]any{
		"slots": slotsResponse,
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
		errors.Is(err, model.ErrInvalidTime):
		return err.Error(), http.StatusBadRequest

	case errors.Is(err, model.ErrNotFound):
		return err.Error(), http.StatusNotFound

	default:
		return "Internal server error", http.StatusInternalServerError
	}
}
