package slot

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type slotService interface {
	GetAvailableSlotsByDate(context.Context, uuid.UUID, time.Time) ([]*model.Slot, error)
}
