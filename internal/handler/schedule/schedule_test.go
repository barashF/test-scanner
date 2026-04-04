package schedule_test

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
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/schedule"
	schedule_controller "github.com/internships-backend/test-backend-barashF/internal/handler/schedule"
	"github.com/internships-backend/test-backend-barashF/internal/handler/schedule/mocks"
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

func TestController_Create(t *testing.T) {
	roomID := uuid.New()

	tests := []struct {
		name           string
		roomIDParam    string
		requestBody    any
		mockBehavior   func(s *mocks.MockscheduleService)
		expectedStatus int
	}{
		{
			name:        "Success",
			roomIDParam: roomID.String(),
			requestBody: dto.CreateRequest{
				DaysOfWeek: []int{1, 3, 5}, // Пн, Ср, Пт
				StartTime:  "09:00",
				EndTime:    "18:00",
			},
			mockBehavior: func(s *mocks.MockscheduleService) {
				s.EXPECT().
					Create(gomock.Any(), roomID, model.DaysOfWeek([]int{1, 3, 5}), "09:00", "18:00").
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid Room UUID in Path",
			roomIDParam:    "not-a-uuid",
			requestBody:    dto.CreateRequest{DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "18:00"},
			mockBehavior:   func(s *mocks.MockscheduleService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON Body",
			roomIDParam:    roomID.String(),
			requestBody:    "invalid-json",
			mockBehavior:   func(s *mocks.MockscheduleService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Conflict (Schedule Exists)",
			roomIDParam: roomID.String(),
			requestBody: dto.CreateRequest{
				DaysOfWeek: []int{1, 2},
				StartTime:  "09:00",
				EndTime:    "18:00",
			},
			mockBehavior: func(s *mocks.MockscheduleService) {
				s.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(model.ErrScheduleExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "Room Not Found",
			roomIDParam: roomID.String(),
			requestBody: dto.CreateRequest{
				DaysOfWeek: []int{1},
				StartTime:  "09:00",
				EndTime:    "18:00",
			},
			mockBehavior: func(s *mocks.MockscheduleService) {
				s.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(model.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "Internal Server Error",
			roomIDParam: roomID.String(),
			requestBody: dto.CreateRequest{
				DaysOfWeek: []int{7},
				StartTime:  "09:00",
				EndTime:    "18:00",
			},
			mockBehavior: func(s *mocks.MockscheduleService) {
				s.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("db connection lost"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockscheduleService(ctrl)
			tt.mockBehavior(mockService)

			handler := schedule_controller.NewController(mockService, noOpLogger{})

			var buf bytes.Buffer
			if s, ok := tt.requestBody.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/rooms/"+tt.roomIDParam+"/schedule", &buf)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("roomId", tt.roomIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d, body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
