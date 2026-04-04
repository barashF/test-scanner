package booking

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type bookingRepository interface {
	Create(context.Context, *model.Booking) (uuid.UUID, error)
	GetByID(context.Context, uuid.UUID) (*model.Booking, error)
	GetAllWithPagination(ctx context.Context, pageSize, offset int) ([]*model.Booking, int, error)
	GetByUserID(context.Context, uuid.UUID) ([]*model.Booking, error)
	UpdateStatus(context.Context, uuid.UUID, model.BookingStatus) error
}

type slotRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Slot, error)
	UpdateIsBooked(ctx context.Context, id uuid.UUID, isBooked bool) error
}

type conferenceGateway interface {
	CreateConference(context.Context, uuid.UUID, time.Time) (string, error)
}

type manager interface {
	InTransaction(ctx context.Context, options *model.TransactionOptions, fn func(context.Context) error) (err error)
}
