package booking_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	booking_controller "github.com/internships-backend/test-backend-barashF/internal/handler/booking"
	"github.com/internships-backend/test-backend-barashF/internal/handler/booking/mocks"
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/booking"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/middleware"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"go.uber.org/mock/gomock"
)

type noOpLogger struct{}

func (n noOpLogger) Debug(m string, f ...logger.Field) {}
func (n noOpLogger) Info(m string, f ...logger.Field)  {}
func (n noOpLogger) Warn(m string, f ...logger.Field)  {}
func (n noOpLogger) Error(m string, f ...logger.Field) {}
func (n noOpLogger) Sync() error                       { return nil }

func withUserContext(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, middleware.UserData{UserID: userID})
	return req.WithContext(ctx)
}

func TestController_Create(t *testing.T) {
	userID := uuid.New()
	slotID := uuid.New()

	tests := []struct {
		name           string
		requestBody    any
		mockBehavior   func(s *mocks.MockbookingService)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: dto.CreateRequest{
				SlotID:               slotID,
				CreateConferenceLink: true,
			},
			mockBehavior: func(s *mocks.MockbookingService) {
				s.EXPECT().
					Create(gomock.Any(), userID, slotID, true).
					Return(&model.Booking{ID: uuid.New(), SlotID: slotID, UserID: userID}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid-json",
			mockBehavior:   func(s *mocks.MockbookingService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Nil Slot ID",
			requestBody: dto.CreateRequest{
				SlotID: uuid.Nil,
			},
			mockBehavior:   func(s *mocks.MockbookingService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Conflict (Already Booked)",
			requestBody: dto.CreateRequest{
				SlotID: slotID,
			},
			mockBehavior: func(s *mocks.MockbookingService) {
				s.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, model.ErrIsBooking)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockbookingService(ctrl)
			tt.mockBehavior(mockService)

			handler := booking_controller.NewController(mockService, noOpLogger{})

			var buf bytes.Buffer
			if s, ok := tt.requestBody.(string); ok {
				buf.WriteString(s)
			} else {
				//nolint:errcheck
				json.NewEncoder(&buf).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/bookings", &buf)
			req = withUserContext(req, userID.String())
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestController_GetAllBookings(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockBehavior   func(s *mocks.MockbookingService)
		expectedStatus int
	}{
		{
			name:  "Success with Defaults",
			query: "",
			mockBehavior: func(s *mocks.MockbookingService) {
				s.EXPECT().
					List(gomock.Any(), 1, 20).
					Return([]*model.Booking{}, 0, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Success with Custom Pagination",
			query: "?page=2&pageSize=10",
			mockBehavior: func(s *mocks.MockbookingService) {
				s.EXPECT().
					List(gomock.Any(), 2, 10).
					Return([]*model.Booking{}, 0, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Page (assuming validation fails on negative or non-digits)",
			query:          "?page=-1",
			mockBehavior:   func(s *mocks.MockbookingService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Service Error",
			query: "",
			mockBehavior: func(s *mocks.MockbookingService) {
				s.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, 0, errors.New("internal"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockbookingService(ctrl)
			tt.mockBehavior(mockService)

			handler := booking_controller.NewController(mockService, noOpLogger{})

			req := httptest.NewRequest(http.MethodGet, "/bookings"+tt.query, nil)
			rr := httptest.NewRecorder()

			handler.GetAllBookings(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestController_GetUserBookings(t *testing.T) {
	userID := uuid.New()

	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockbookingService(ctrl)

	mockService.EXPECT().
		GetByUser(gomock.Any(), userID).
		Return([]*model.Booking{{ID: uuid.New()}}, nil)

	handler := booking_controller.NewController(mockService, noOpLogger{})

	req := httptest.NewRequest(http.MethodGet, "/bookings/my", nil)
	req = withUserContext(req, userID.String())
	rr := httptest.NewRecorder()

	handler.GetUserBookings(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestController_CancelBooking(t *testing.T) {
	userID := uuid.New()
	bookingID := uuid.New()

	tests := []struct {
		name           string
		bookingIDParam string
		mockBehavior   func(s *mocks.MockbookingService)
		expectedStatus int
	}{
		{
			name:           "Success",
			bookingIDParam: bookingID.String(),
			mockBehavior: func(s *mocks.MockbookingService) {
				s.EXPECT().
					Cancel(gomock.Any(), bookingID, userID).
					Return(&model.Booking{ID: bookingID}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid UUID in Path",
			bookingIDParam: "not-a-uuid",
			mockBehavior:   func(s *mocks.MockbookingService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Forbidden (Not Owner)",
			bookingIDParam: bookingID.String(),
			mockBehavior: func(s *mocks.MockbookingService) {
				s.EXPECT().
					Cancel(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, model.ErrBookingOwnershipRequired)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Not Found",
			bookingIDParam: bookingID.String(),
			mockBehavior: func(s *mocks.MockbookingService) {
				s.EXPECT().
					Cancel(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, model.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockbookingService(ctrl)
			tt.mockBehavior(mockService)

			handler := booking_controller.NewController(mockService, noOpLogger{})

			req := httptest.NewRequest(http.MethodDelete, "/bookings/"+tt.bookingIDParam, nil)
			req = withUserContext(req, userID.String())
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("bookingId", tt.bookingIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.CancelBooking(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d, body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
