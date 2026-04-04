package validation_test

import (
	"testing"
	"time"

	"github.com/internships-backend/test-backend-barashF/internal/handler/slot/validation"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func TestValidateSlotDate(t *testing.T) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)
	exactlyOneYear := today.AddDate(1, 0, 0)
	overOneYear := today.AddDate(1, 0, 1)

	tests := []struct {
		name        string
		dateStr     string
		expectedErr error
		checkDate   bool
		expectedDay time.Time
	}{
		{
			name:        "Success Today",
			dateStr:     today.Format("2006-01-02"),
			expectedErr: nil,
			checkDate:   true,
			expectedDay: today,
		},
		{
			name:        "Success Tomorrow",
			dateStr:     tomorrow.Format("2006-01-02"),
			expectedErr: nil,
			checkDate:   true,
			expectedDay: tomorrow,
		},
		{
			name:        "Success Exactly One Year From Now",
			dateStr:     exactlyOneYear.Format("2006-01-02"),
			expectedErr: nil,
			checkDate:   true,
			expectedDay: exactlyOneYear,
		},
		{
			name:        "Empty Date String",
			dateStr:     "",
			expectedErr: model.ErrMissingRequiredFields,
		},
		{
			name:        "Invalid Date Format",
			dateStr:     "04-04-2026",
			expectedErr: model.ErrInvalidTime,
		},
		{
			name:        "Garbage String",
			dateStr:     "not-a-date",
			expectedErr: model.ErrInvalidTime,
		},
		{
			name:        "Date In The Past (Yesterday)",
			dateStr:     yesterday.Format("2006-01-02"),
			expectedErr: model.ErrInvalidTime,
		},
		{
			name:        "Date Too Far In The Future (Over One Year)",
			dateStr:     overOneYear.Format("2006-01-02"),
			expectedErr: model.ErrInvalidTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedDate, err := validation.ValidateSlotDate(tt.dateStr)

			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.checkDate && err == nil {
				if !parsedDate.Equal(tt.expectedDay) {
					t.Errorf("expected parsed date %v, got %v", tt.expectedDay, parsedDate)
				}
			}
		})
	}
}
