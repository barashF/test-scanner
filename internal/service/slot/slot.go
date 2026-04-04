package slot

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type Service struct {
	slotRepo slotRepository
	roomRepo roomRepository
	logger   logger.Logger
}

func NewService(
	slotRepo slotRepository,
	roomRepo roomRepository,
	logger logger.Logger,
) *Service {
	return &Service{
		slotRepo: slotRepo,
		roomRepo: roomRepo,
		logger:   logger,
	}
}

func (s *Service) GetAvailableSlotsByDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]*model.Slot, error) {
	_, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		s.logger.Warn("Failed to get room for fetching slots",
			logger.F("room_id", roomID),
			logger.F("error", err.Error()),
		)
		return nil, err
	}

	slots, err := s.slotRepo.GetByRoomAndDate(ctx, roomID, date)
	if err != nil {
		s.logger.Error("Database error when fetching available slots",
			logger.F("room_id", roomID),
			logger.F("date", date.Format("2006-01-02")),
			logger.F("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to fetch slots from repository: %w", err)
	}

	s.logger.Info("Available slots fetched successfully",
		logger.F("room_id", roomID),
		logger.F("date", date.Format("2006-01-02")),
		logger.F("slots_found", len(slots)),
	)

	return slots, nil
}
