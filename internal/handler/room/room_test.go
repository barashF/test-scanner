package room_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/room"
	room_controller "github.com/internships-backend/test-backend-barashF/internal/handler/room"
	"github.com/internships-backend/test-backend-barashF/internal/handler/room/mocks"
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
	tests := []struct {
		name           string
		requestBody    any
		mockBehavior   func(s *mocks.MockroomService)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: dto.CreateRequest{
				Name:        "Conference Room A",
				Description: "Equipped with projector",
				Capacity:    10,
			},
			mockBehavior: func(s *mocks.MockroomService) {
				s.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx any, room *model.Room) (uuid.UUID, error) {
						if room.Name != "Conference Room A" || room.Capacity != 10 {
							return uuid.Nil, errors.New("unexpected room data")
						}
						return uuid.New(), nil
					})
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid-json",
			mockBehavior:   func(s *mocks.MockroomService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			requestBody: dto.CreateRequest{
				Name:        "Conference Room B",
				Description: "Small room",
				Capacity:    4,
			},
			mockBehavior: func(s *mocks.MockroomService) {
				s.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockroomService(ctrl)
			tt.mockBehavior(mockService)

			handler := room_controller.NewController(mockService, noOpLogger{})

			var buf bytes.Buffer
			if s, ok := tt.requestBody.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/rooms", &buf)
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestController_List(t *testing.T) {
	tests := []struct {
		name           string
		mockBehavior   func(s *mocks.MockroomService)
		expectedStatus int
	}{
		{
			name: "Success with Data",
			mockBehavior: func(s *mocks.MockroomService) {
				rooms := []*model.Room{
					{ID: uuid.New(), Name: "Room 1", Capacity: 5, CreatedAt: time.Now().UTC()},
					{ID: uuid.New(), Name: "Room 2", Capacity: 15, CreatedAt: time.Now().UTC()},
				}
				s.EXPECT().List(gomock.Any()).Return(rooms, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Success Empty",
			mockBehavior: func(s *mocks.MockroomService) {
				s.EXPECT().List(gomock.Any()).Return([]*model.Room{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Service Error",
			mockBehavior: func(s *mocks.MockroomService) {
				s.EXPECT().List(gomock.Any()).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockroomService(ctrl)
			tt.mockBehavior(mockService)

			handler := room_controller.NewController(mockService, noOpLogger{})

			req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
			rr := httptest.NewRecorder()

			handler.List(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
