package schedule_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/internships-backend/test-backend-barashF/internal/service/schedule"
	"github.com/internships-backend/test-backend-barashF/internal/service/schedule/mocks"
	"go.uber.org/mock/gomock"
)

func TestCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockScheduleRepo := mocks.NewMockscheduleRepository(ctrl)
	mockRoomRepo := mocks.NewMockroomRepository(ctrl)
	mockSlotRepo := mocks.NewMockslotRepository(ctrl)
	mockManager := mocks.NewMockmanager(ctrl)

	s := schedule.NewService(mockScheduleRepo, mockRoomRepo, mockSlotRepo, mockManager, noOpLogger{})

	roomID := uuid.New()
	days := model.DaysOfWeek{1, 2, 3}

	mockRoomRepo.EXPECT().GetByID(gomock.Any(), roomID).Return(&model.Room{ID: roomID}, nil)
	mockScheduleRepo.EXPECT().GetByRoomID(gomock.Any(), roomID).Return(nil, model.ErrNotFound)

	mockManager.EXPECT().
		InTransaction(gomock.Any(), gomock.Nil(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts *model.TransactionOptions, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockScheduleRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(uuid.New(), nil)
	mockSlotRepo.EXPECT().CreateMany(gomock.Any(), gomock.Any()).Return(nil)

	err := s.Create(context.Background(), roomID, days, "09:00", "18:00")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreate_RoomNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRoomRepo := mocks.NewMockroomRepository(ctrl)

	s := schedule.NewService(nil, mockRoomRepo, nil, nil, noOpLogger{})

	roomID := uuid.New()
	dbErr := errors.New("room not found")

	mockRoomRepo.EXPECT().GetByID(gomock.Any(), roomID).Return(nil, dbErr)

	err := s.Create(context.Background(), roomID, model.DaysOfWeek{1}, "09:00", "18:00")
	if !errors.Is(err, dbErr) {
		t.Errorf("expected error %v, got %v", dbErr, err)
	}
}

func TestCreate_ScheduleAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockScheduleRepo := mocks.NewMockscheduleRepository(ctrl)
	mockRoomRepo := mocks.NewMockroomRepository(ctrl)

	s := schedule.NewService(mockScheduleRepo, mockRoomRepo, nil, nil, noOpLogger{})

	roomID := uuid.New()

	mockRoomRepo.EXPECT().GetByID(gomock.Any(), roomID).Return(&model.Room{ID: roomID}, nil)
	mockScheduleRepo.EXPECT().GetByRoomID(gomock.Any(), roomID).Return(&model.Schedule{}, nil)

	err := s.Create(context.Background(), roomID, model.DaysOfWeek{1}, "09:00", "18:00")
	if !errors.Is(err, model.ErrScheduleExists) {
		t.Errorf("expected ErrScheduleExists, got %v", err)
	}
}

func TestCreate_InvalidTimeFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockScheduleRepo := mocks.NewMockscheduleRepository(ctrl)
	mockRoomRepo := mocks.NewMockroomRepository(ctrl)

	s := schedule.NewService(mockScheduleRepo, mockRoomRepo, nil, nil, noOpLogger{})

	roomID := uuid.New()

	mockRoomRepo.EXPECT().GetByID(gomock.Any(), roomID).Return(&model.Room{ID: roomID}, nil)
	mockScheduleRepo.EXPECT().GetByRoomID(gomock.Any(), roomID).Return(nil, model.ErrNotFound)

	err := s.Create(context.Background(), roomID, model.DaysOfWeek{1}, "invalid_time", "18:00")
	if !errors.Is(err, model.ErrInvalidTime) {
		t.Errorf("expected ErrInvalidTime, got %v", err)
	}
}

func TestCreate_TransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockScheduleRepo := mocks.NewMockscheduleRepository(ctrl)
	mockRoomRepo := mocks.NewMockroomRepository(ctrl)
	mockManager := mocks.NewMockmanager(ctrl)

	s := schedule.NewService(mockScheduleRepo, mockRoomRepo, nil, mockManager, noOpLogger{})

	roomID := uuid.New()
	txErr := errors.New("transaction failed")

	mockRoomRepo.EXPECT().GetByID(gomock.Any(), roomID).Return(&model.Room{ID: roomID}, nil)
	mockScheduleRepo.EXPECT().GetByRoomID(gomock.Any(), roomID).Return(nil, model.ErrNotFound)

	mockManager.EXPECT().
		InTransaction(gomock.Any(), gomock.Nil(), gomock.Any()).
		Return(txErr)

	err := s.Create(context.Background(), roomID, model.DaysOfWeek{1}, "09:00", "18:00")
	if !errors.Is(err, txErr) {
		t.Errorf("expected error %v, got %v", txErr, err)
	}
}

type noOpLogger struct{}

func (n noOpLogger) Debug(msg string, fields ...logger.Field) {}
func (n noOpLogger) Info(msg string, fields ...logger.Field)  {}
func (n noOpLogger) Warn(msg string, fields ...logger.Field)  {}
func (n noOpLogger) Error(msg string, fields ...logger.Field) {}
func (n noOpLogger) Sync() error                              { return nil }
