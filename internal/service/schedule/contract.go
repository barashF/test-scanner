package schedule

import (
	"context"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type scheduleRepository interface {
	GetByRoomID(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error)
	Create(context.Context, *model.Schedule) (uuid.UUID, error)
}

type roomRepository interface {
	GetByID(context.Context, uuid.UUID) (*model.Room, error)
}

type slotRepository interface {
	CreateMany(context.Context, []*model.Slot) error
}

type manager interface {
	InTransaction(ctx context.Context, options *model.TransactionOptions, fn func(context.Context) error) (err error)
}
