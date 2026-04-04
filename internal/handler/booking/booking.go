package booking

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/handler/booking/validation"
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/booking"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/middleware"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type Controller struct {
	serviceBooking bookingService
	logger         logger.Logger
}

func NewController(service bookingService, logger logger.Logger) *Controller {
	return &Controller{
		serviceBooking: service,
		logger:         logger,
	}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(middleware.UserContextKey).(middleware.UserData)
	userID, _ := uuid.Parse(userData.UserID)

	var req dto.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	if req.SlotID == uuid.Nil {
		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	booking, err := c.serviceBooking.Create(r.Context(), userID, req.SlotID, req.CreateConferenceLink)
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	c.writeJSON(w, http.StatusCreated, map[string]any{
		"booking": dto.ModelToResponse(booking),
	})
}

func (c *Controller) GetAllBookings(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("pageSize")

	if err := validation.ValidatePage(pageStr); err != nil {
		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid pagination",
		})
		return
	}

	if err := validation.ValidateSize(sizeStr); err != nil {
		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid pagination",
		})
		return
	}

	page := validation.Convert(pageStr, 1)
	size := validation.Convert(sizeStr, 20)

	bookings, totalCount, err := c.serviceBooking.List(r.Context(), page, size)
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	bookingsResponse := make([]dto.Response, 0, len(bookings))
	for _, booking := range bookings {
		bookingsResponse = append(bookingsResponse, dto.ModelToResponse(booking))
	}

	c.writeJSON(w, http.StatusOK, map[string]any{
		"bookings":    bookingsResponse,
		"total_count": totalCount,
	})
}

func (c *Controller) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(middleware.UserContextKey).(middleware.UserData)
	userID, _ := uuid.Parse(userData.UserID)

	bookings, err := c.serviceBooking.GetByUser(r.Context(), userID)
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	bookingsResponse := make([]dto.Response, 0, len(bookings))
	for _, booking := range bookings {
		bookingsResponse = append(bookingsResponse, dto.ModelToResponse(booking))
	}

	c.writeJSON(w, http.StatusOK, map[string]any{
		"bookings": bookingsResponse,
	})
}

func (c *Controller) CancelBooking(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(middleware.UserContextKey).(middleware.UserData)
	userID, _ := uuid.Parse(userData.UserID)

	bookingIdStr := chi.URLParam(r, "bookingId")
	bookingID, err := uuid.Parse(bookingIdStr)
	if err != nil {
		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid booking ID in path",
		})
		return
	}

	booking, err := c.serviceBooking.Cancel(r.Context(), bookingID, userID)
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	c.writeJSON(w, http.StatusOK, map[string]any{
		"booking": dto.ModelToResponse(booking),
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

	case errors.Is(err, model.ErrBookingOwnershipRequired):
		return err.Error(), http.StatusForbidden

	case errors.Is(err, model.ErrNotFound):
		return err.Error(), http.StatusNotFound

	case errors.Is(err, model.ErrIsBooking):
		return "slot is already booked", http.StatusConflict

	default:
		return "Internal server error", http.StatusInternalServerError
	}
}
