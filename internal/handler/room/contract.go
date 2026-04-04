package room

import (
	"context"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type roomService interface {
	Create(context.Context, *model.Room) (uuid.UUID, error)
	List(context.Context) ([]*model.Room, error)
}
