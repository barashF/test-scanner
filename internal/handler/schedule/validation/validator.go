package validation

import (
	"fmt"
	"time"

	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/schedule"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func ValidateCreateRequestSchedule(req *dto.CreateRequest) error {
	err := validateDaysOfWeek(req.DaysOfWeek)
	if err != nil {
		return err
	}

	err = validateTime(req.StartTime, req.EndTime)
	if err != nil {
		return err
	}

	return nil
}

func validateDaysOfWeek(days []int) error {
	if len(days) == 0 {
		return model.ErrInvalidDaysOfWeek
	}

	for _, day := range days {
		if day < 1 || day > 7 {
			return model.ErrInvalidDaysOfWeek
		}
	}

	return nil
}

func validateTime(startTime, endTime string) error {
	if startTime == "" || endTime == "" {
		return model.ErrInvalidTime
	}

	start, err := time.Parse("15:04", startTime)
	if err != nil {
		return fmt.Errorf("invalid start time format (must be HH:MM): %w", model.ErrInvalidTime)
	}

	end, err := time.Parse("15:04", endTime)
	if err != nil {
		return fmt.Errorf("invalid end time format (must be HH:MM): %w", model.ErrInvalidTime)
	}

	if !start.Before(end) {
		return fmt.Errorf("start time must be before end time: %w", model.ErrInvalidTime)
	}

	return nil
}
