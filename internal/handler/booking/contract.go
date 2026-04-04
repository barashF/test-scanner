package booking

import (
	"context"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type bookingService interface {
	Create(ctx context.Context, userID uuid.UUID, slotID uuid.UUID, createConferenceLink bool) (*model.Booking, error)
	List(ctx context.Context, page, pageSize int) ([]*model.Booking, int, error)
	GetByUser(context.Context, uuid.UUID) ([]*model.Booking, error)
	Cancel(ctx context.Context, bookingID, userID uuid.UUID) (*model.Booking, error)
}
