package slot_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/internships-backend/test-backend-barashF/internal/service/slot"
	"github.com/internships-backend/test-backend-barashF/internal/service/slot/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetAvailableSlotsByDate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSlotRepo := mocks.NewMockslotRepository(ctrl)
	mockRoomRepo := mocks.NewMockroomRepository(ctrl)

	s := slot.NewService(mockSlotRepo, mockRoomRepo, noOpLogger{})

	roomID := uuid.New()
	date := time.Now()

	mockSlots := []*model.Slot{
		{ID: uuid.New(), RoomID: roomID},
		{ID: uuid.New(), RoomID: roomID},
	}

	mockRoomRepo.EXPECT().GetByID(gomock.Any(), roomID).Return(&model.Room{ID: roomID}, nil)
	mockSlotRepo.EXPECT().GetByRoomAndDate(gomock.Any(), roomID, date).Return(mockSlots, nil)

	slots, err := s.GetAvailableSlotsByDate(context.Background(), roomID, date)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(slots) != 2 {
		t.Errorf("expected 2 slots, got %d", len(slots))
	}
}

func TestGetAvailableSlotsByDate_RoomNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRoomRepo := mocks.NewMockroomRepository(ctrl)

	s := slot.NewService(nil, mockRoomRepo, noOpLogger{})

	roomID := uuid.New()
	dbErr := errors.New("room not found")

	mockRoomRepo.EXPECT().GetByID(gomock.Any(), roomID).Return(nil, dbErr)

	_, err := s.GetAvailableSlotsByDate(context.Background(), roomID, time.Now())
	if !errors.Is(err, dbErr) {
		t.Errorf("expected error %v, got %v", dbErr, err)
	}
}

func TestGetAvailableSlotsByDate_SlotRepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSlotRepo := mocks.NewMockslotRepository(ctrl)
	mockRoomRepo := mocks.NewMockroomRepository(ctrl)

	s := slot.NewService(mockSlotRepo, mockRoomRepo, noOpLogger{})

	roomID := uuid.New()
	date := time.Now()
	dbErr := errors.New("db error")

	mockRoomRepo.EXPECT().GetByID(gomock.Any(), roomID).Return(&model.Room{ID: roomID}, nil)
	mockSlotRepo.EXPECT().GetByRoomAndDate(gomock.Any(), roomID, date).Return(nil, dbErr)

	_, err := s.GetAvailableSlotsByDate(context.Background(), roomID, date)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Errorf("expected wrapped error containing %v, got %v", dbErr, err)
	}
}

type noOpLogger struct{}

func (n noOpLogger) Debug(msg string, fields ...logger.Field) {}
func (n noOpLogger) Info(msg string, fields ...logger.Field)  {}
func (n noOpLogger) Warn(msg string, fields ...logger.Field)  {}
func (n noOpLogger) Error(msg string, fields ...logger.Field) {}
func (n noOpLogger) Sync() error                              { return nil }
