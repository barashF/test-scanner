package slot_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	slot_controller "github.com/internships-backend/test-backend-barashF/internal/handler/slot"
	"github.com/internships-backend/test-backend-barashF/internal/handler/slot/mocks"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"go.uber.org/mock/gomock"
)

type noOpLogger struct{}

func (n noOpLogger) Debug(m string, f ...logger.Field) {}
func (n noOpLogger) Info(m string, f ...logger.Field)  {}
func (n noOpLogger) Warn(m string, f ...logger.Field)  {}
func (n noOpLogger) Error(m string, f ...logger.Field) {}
func (n noOpLogger) Sync() error                       { return nil }

func TestController_GetAvailableSlots(t *testing.T) {
	roomID := uuid.New()

	today := time.Now().UTC().Format("2006-01-02")
	fixedDate, _ := time.Parse("2006-01-02", today)

	tests := []struct {
		name           string
		roomIDParam    string
		query          string
		mockBehavior   func(s *mocks.MockslotService)
		expectedStatus int
	}{
		{
			name:        "Success",
			roomIDParam: roomID.String(),
			query:       fmt.Sprintf("?date=%s", today),
			mockBehavior: func(s *mocks.MockslotService) {
				s.EXPECT().
					GetAvailableSlotsByDate(gomock.Any(), roomID, fixedDate).
					Return([]*model.Slot{
						{ID: uuid.New(), RoomID: roomID, Start: fixedDate, End: fixedDate.Add(1 * time.Hour)},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Room UUID",
			roomIDParam:    "not-a-uuid",
			query:          fmt.Sprintf("?date=%s", today),
			mockBehavior:   func(s *mocks.MockslotService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Service Error",
			roomIDParam: roomID.String(),
			query:       fmt.Sprintf("?date=%s", today),
			mockBehavior: func(s *mocks.MockslotService) {
				s.EXPECT().
					GetAvailableSlotsByDate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database failure"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "Slots Not Found",
			roomIDParam: roomID.String(),
			query:       fmt.Sprintf("?date=%s", today),
			mockBehavior: func(s *mocks.MockslotService) {
				s.EXPECT().
					GetAvailableSlotsByDate(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, model.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockslotService(ctrl)
			tt.mockBehavior(mockService)

			handler := slot_controller.NewController(mockService, noOpLogger{})

			req := httptest.NewRequest(http.MethodGet, "/rooms/"+tt.roomIDParam+"/slots"+tt.query, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("roomId", tt.roomIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.GetAvailableSlots(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d, body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
