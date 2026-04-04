package room

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/room"
	"github.com/internships-backend/test-backend-barashF/internal/handler/room/validation"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type Controller struct {
	roomService roomService
	logger      logger.Logger
}

func NewController(service roomService, logger logger.Logger) *Controller {
	return &Controller{
		roomService: service,
		logger:      logger,
	}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	err := validation.ValidateCreateRequestRoom(&req)
	if err != nil {
		c.logger.Warn("Create room request validation failed",
			logger.F("name", req.Name),
			logger.F("error", err.Error()),
		)
		c.writeError(w, err, r)
		return
	}

	room := &model.Room{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
		CreatedAt:   time.Now().UTC(),
	}

	roomID, err := c.roomService.Create(r.Context(), room)
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	c.writeJSON(w, http.StatusCreated, map[string]any{
		"id": roomID,
	})
}

func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	rooms, err := c.roomService.List(r.Context())
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	roomsResponse := make([]dto.RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		roomsResponse = append(roomsResponse, dto.ModelToResponse(room))
	}

	c.writeJSON(w, http.StatusOK, map[string]any{
		"rooms": roomsResponse,
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
		errors.Is(err, model.ErrValidationFailed):
		return err.Error(), http.StatusBadRequest

	default:
		return "Internal server error", http.StatusInternalServerError
	}
}
