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

// Create godoc
// @Summary      Создать бронь на слот (только user).
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body      dto.CreateRequest  true  "Данные для создания бронирования"
// @Success      201  {object}  dto.BookingCreateResponse "Бронь создана"
// @Failure      400  {object}  dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object}  dto.ErrorResponse "Не авторизован"
// @Failure      403  {object}  dto.ErrorResponse "Доступ запрещён (бронирование доступно только роли user)"
// @Failure      404  {object}  dto.ErrorResponse "Слот не найден"
// @Failure      409  {object}  dto.ErrorResponse "Слот уже занят"
// @Failure      500  {object}  dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /bookings/create [post]
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

	c.writeJSON(w, http.StatusCreated, dto.BookingCreateResponse{
		Booking: dto.ModelToResponse(booking),
	})
}

// GetAllBookings godoc
// @Summary      Список всех броней с пагинацией (только admin)
// @Description  Доступно только роли admin. Поддерживает пагинацию.
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int  false  "Номер страницы (начиная с 1). По умолчанию 1." minimum(1) default(1)
// @Param        pageSize  query     int  false  "Количество записей на странице. По умолчанию 20, максимум 100." minimum(1) maximum(100) default(20)
// @Success      200  {object}  dto.BookingListPaginatationResponse "Список всех броней и метаданные пагинации"
// @Failure      400  {object}  dto.ErrorResponse "Неверный запрос (некорректные параметры пагинации)"
// @Failure      401  {object}  dto.ErrorResponse "Не авторизован"
// @Failure      403  {object}  dto.ErrorResponse "Доступ запрещён (только admin)"
// @Failure      500  {object}  dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /bookings/list [get]
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

// GetUserBookings godoc
// @Summary      Список броней текущего пользователя (только user)
// @Description  Доступно только роли user. Возвращает брони пользователя.
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.BookingListResponse "Список броней текущего пользователя"
// @Failure      401  {object}  dto.ErrorResponse "Не авторизован"
// @Failure      403  {object}  dto.ErrorResponse "Доступ запрещён (только user)"
// @Failure      500  {object}  dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /bookings/my [get]
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

// CancelBooking godoc
// @Summary      Отменить бронь (только своя бронь, только user)
// @Description  Доступно только роли user. Пользователь может отменить только свою бронь.
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        bookingId  path      string  true  "Идентификатор брони (UUID)" format(uuid)
// @Success      200  {object}  dto.BookingCreateResponse "Бронь отменена (или уже была отменена ранее)"
// @Failure      400  {object}  dto.ErrorResponse "Неверный формат ID брони"
// @Failure      401  {object}  dto.ErrorResponse "Не авторизован"
// @Failure      403  {object}  dto.ErrorResponse "Не своя бронь или роль не user"
// @Failure      404  {object}  dto.ErrorResponse "Бронь не найдена"
// @Failure      500  {object}  dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /bookings/{bookingId}/cancel [post]
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
