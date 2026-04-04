package schedule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

const (
	daysToGenerate = 30
	slotDuration   = 30 * time.Minute
)

type Service struct {
	scheduleRepo scheduleRepository
	roomRepo     roomRepository
	slotRepo     slotRepository
	manager      manager
	logger       logger.Logger
}

func NewService(
	scheduleRepo scheduleRepository,
	roomRepo roomRepository,
	slotRepo slotRepository,
	manager manager,
	logger logger.Logger,
) *Service {
	return &Service{
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
		slotRepo:     slotRepo,
		manager:      manager,
		logger:       logger,
	}
}

func (s *Service) Create(ctx context.Context, roomID uuid.UUID, daysOfWeek model.DaysOfWeek, startTimeStr, endTimeStr string) error {
	_, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		s.logger.Warn("Failed to get room for schedule creation",
			logger.F("room_id", roomID),
			logger.F("error", err.Error()),
		)
		return err
	}

	scheduleExists, err := s.scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		s.logger.Error("Database error when checking existing schedule",
			logger.F("room_id", roomID),
			logger.F("error", err.Error()),
		)
		return err
	}
	if scheduleExists != nil {
		s.logger.Warn("Schedule already exists for the room", logger.F("room_id", roomID))
		return model.ErrScheduleExists
	}

	slots, err := s.generateSlots(daysOfWeek, roomID, startTimeStr, endTimeStr)
	if err != nil {
		s.logger.Warn("Invalid time range provided for schedule",
			logger.F("room_id", roomID),
			logger.F("start_time", startTimeStr),
			logger.F("end_time", endTimeStr),
			logger.F("error", err),
		)
		return model.ErrInvalidTime
	}

	schedule := &model.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  startTimeStr,
		EndTime:    endTimeStr,
	}

	err = s.manager.InTransaction(ctx, nil, func(dbContext context.Context) error {
		_, err := s.scheduleRepo.Create(dbContext, schedule)
		if err != nil {
			return err
		}

		err = s.slotRepo.CreateMany(dbContext, slots)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.logger.Error("Failed to save schedule and slots within transaction",
			logger.F("room_id", roomID),
			logger.F("error", err.Error()),
		)
		return err
	}

	s.logger.Info("Schedule created successfully",
		logger.F("room_id", roomID),
		logger.F("slots_generated", len(slots)),
	)

	return nil
}

func (s *Service) generateSlots(daysOfWeek model.DaysOfWeek, roomID uuid.UUID, startTimeStr, endTimeStr string) ([]*model.Slot, error) {
	startHour, startMin, err := parseTime(startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}
	endHour, endMin, err := parseTime(endTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	var slots []*model.Slot
	now := time.Now().UTC()

	for i := 0; i < daysToGenerate; i++ {
		currentDate := now.AddDate(0, 0, i)

		if !isDayAllowed(currentDate.Weekday(), daysOfWeek) {
			continue
		}

		startSlotTime := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), startHour, startMin, 0, 0, time.UTC)
		endWorkTime := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), endHour, endMin, 0, 0, time.UTC)

		for startSlotTime.Before(endWorkTime) {
			endSlotTime := startSlotTime.Add(slotDuration)

			if endSlotTime.After(endWorkTime) {
				break
			}

			slots = append(slots, &model.Slot{
				ID:     uuid.New(),
				RoomID: roomID,
				Start:  startSlotTime,
				End:    endSlotTime,
			})

			startSlotTime = endSlotTime
		}
	}

	return slots, nil
}

func parseTime(tStr string) (hour, min int, err error) {
	_, err = fmt.Sscanf(tStr, "%d:%d", &hour, &min)
	if err != nil {
		return 0, 0, err
	}
	return hour, min, nil
}

func isDayAllowed(day time.Weekday, allowedDays model.DaysOfWeek) bool {
	currentDayInt := int(day)
	if currentDayInt == 0 {
		currentDayInt = 7
	}

	for _, allowed := range allowedDays {
		if allowed == currentDayInt {
			return true
		}
	}

	return false
}
