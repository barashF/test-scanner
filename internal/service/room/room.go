package room

import (
	"context"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type Service struct {
	roomRepo roomRepository
}

func NewService(roomRepo roomRepository) *Service {
	return &Service{
		roomRepo: roomRepo,
	}
}

func (s *Service) Create(ctx context.Context, room *model.Room) (uuid.UUID, error) {
	_, err := s.roomRepo.Create(ctx, room)
	if err != nil {
		return uuid.Nil, err
	}

	return room.ID, nil
}

func (s *Service) List(ctx context.Context) ([]*model.Room, error) {
	rooms, err := s.roomRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	room, err := s.roomRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return room, nil
}
