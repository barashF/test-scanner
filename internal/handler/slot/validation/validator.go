package validation

import (
	"time"

	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func ValidateSlotDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, model.ErrMissingRequiredFields
	}

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, model.ErrInvalidTime
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if parsedDate.Before(today) {
		return time.Time{}, model.ErrInvalidTime
	}

	maxFutureDate := today.AddDate(1, 0, 0)
	if parsedDate.After(maxFutureDate) {
		return time.Time{}, model.ErrInvalidTime
	}

	return parsedDate, nil
}
