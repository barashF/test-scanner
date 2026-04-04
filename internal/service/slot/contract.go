package slot

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type slotRepository interface {
	GetByRoomAndDate(context.Context, uuid.UUID, time.Time) ([]*model.Slot, error)
}

type roomRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error)
}
