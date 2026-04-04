package schedule

import (
	"context"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type scheduleService interface {
	Create(ctx context.Context, roomID uuid.UUID, daysOfWeek model.DaysOfWeek, startTimeStr, endTimeStr string) error
}
