package room

import (
	"context"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type roomRepository interface {
	GetByID(context.Context, uuid.UUID) (*model.Room, error)
	Create(context.Context, *model.Room) (uuid.UUID, error)
	GetAll(context.Context) ([]*model.Room, error)
}
