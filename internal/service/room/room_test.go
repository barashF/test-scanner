package room_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/internships-backend/test-backend-barashF/internal/service/room"
	"github.com/internships-backend/test-backend-barashF/internal/service/room/mocks"

	"go.uber.org/mock/gomock"
)

func TestCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockroomRepository(ctrl)

	s := room.NewService(mockRepo)

	roomModel := &model.Room{
		ID:   uuid.New(),
		Name: "Confrere Room 1",
	}

	mockRepo.EXPECT().
		Create(gomock.Any(), roomModel).
		Return(roomModel.ID, nil).
		Times(1)

	id, err := s.Create(context.Background(), roomModel)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != roomModel.ID {
		t.Errorf("expected id %v, got %v", roomModel.ID, id)
	}
}

func TestCreate_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockroomRepository(ctrl)

	s := room.NewService(mockRepo)

	roomModel := &model.Room{ID: uuid.New()}
	dbErr := errors.New("database connection failed")

	mockRepo.EXPECT().
		Create(gomock.Any(), roomModel).
		Return(uuid.Nil, dbErr).
		Times(1)

	_, err := s.Create(context.Background(), roomModel)

	if !errors.Is(err, dbErr) {
		t.Errorf("expected error %v, got %v", dbErr, err)
	}
}

func TestList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockroomRepository(ctrl)

	s := room.NewService(mockRepo)

	mockRooms := []*model.Room{
		{ID: uuid.New(), Name: "Room A"},
		{ID: uuid.New(), Name: "Room B"},
	}

	mockRepo.EXPECT().
		GetAll(gomock.Any()).
		Return(mockRooms, nil).
		Times(1)

	rooms, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(rooms))
	}
}

func TestList_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockroomRepository(ctrl)

	s := room.NewService(mockRepo)

	dbErr := errors.New("db error")

	mockRepo.EXPECT().
		GetAll(gomock.Any()).
		Return(nil, dbErr).
		Times(1)

	_, err := s.List(context.Background())

	if !errors.Is(err, dbErr) {
		t.Errorf("expected error %v, got %v", dbErr, err)
	}
}

func TestGetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockroomRepository(ctrl)

	s := room.NewService(mockRepo)

	roomID := uuid.New()
	mockRoom := &model.Room{ID: roomID, Name: "Target Room"}

	mockRepo.EXPECT().
		GetByID(gomock.Any(), roomID).
		Return(mockRoom, nil).
		Times(1)

	r, err := s.GetByID(context.Background(), roomID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.ID != roomID {
		t.Errorf("expected room ID %v, got %v", roomID, r.ID)
	}
}

func TestGetByID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockroomRepository(ctrl)

	s := room.NewService(mockRepo)

	roomID := uuid.New()
	dbErr := errors.New("room not found")

	mockRepo.EXPECT().
		GetByID(gomock.Any(), roomID).
		Return(nil, dbErr).
		Times(1)

	_, err := s.GetByID(context.Background(), roomID)

	if !errors.Is(err, dbErr) {
		t.Errorf("expected error %v, got %v", dbErr, err)
	}
}
