package validation_test

import (
	"errors"
	"testing"

	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/schedule"
	"github.com/internships-backend/test-backend-barashF/internal/handler/schedule/validation"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func TestValidateCreateRequestSchedule(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.CreateRequest
		expectedErr error
	}{
		{
			name: "Success",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{1, 3, 5},
				StartTime:  "09:00",
				EndTime:    "18:00",
			},
			expectedErr: nil,
		},
		{
			name: "Success Boundary Time",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{1},
				StartTime:  "00:00",
				EndTime:    "23:59",
			},
			expectedErr: nil,
		},
		{
			name: "Empty Days Of Week",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{},
				StartTime:  "09:00",
				EndTime:    "18:00",
			},
			expectedErr: model.ErrInvalidDaysOfWeek,
		},
		{
			name: "Nil Days Of Week",
			req: &dto.CreateRequest{
				DaysOfWeek: nil,
				StartTime:  "09:00",
				EndTime:    "18:00",
			},
			expectedErr: model.ErrInvalidDaysOfWeek,
		},
		{
			name: "Invalid Day Number (Too Low)",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{0, 2},
				StartTime:  "09:00",
				EndTime:    "18:00",
			},
			expectedErr: model.ErrInvalidDaysOfWeek,
		},
		{
			name: "Invalid Day Number (Too High)",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{1, 8},
				StartTime:  "09:00",
				EndTime:    "18:00",
			},
			expectedErr: model.ErrInvalidDaysOfWeek,
		},
		{
			name: "Empty Start Time",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{1},
				StartTime:  "",
				EndTime:    "18:00",
			},
			expectedErr: model.ErrInvalidTime,
		},
		{
			name: "Empty End Time",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{1},
				StartTime:  "09:00",
				EndTime:    "",
			},
			expectedErr: model.ErrInvalidTime,
		},
		{
			name: "Invalid End Time Format",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{1},
				StartTime:  "09:00",
				EndTime:    "18-00",
			},
			expectedErr: model.ErrInvalidTime,
		},
		{
			name: "Start Time Equals End Time",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{1},
				StartTime:  "09:00",
				EndTime:    "09:00",
			},
			expectedErr: model.ErrInvalidTime,
		},
		{
			name: "Start Time After End Time",
			req: &dto.CreateRequest{
				DaysOfWeek: []int{1},
				StartTime:  "18:00",
				EndTime:    "09:00",
			},
			expectedErr: model.ErrInvalidTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateCreateRequestSchedule(tt.req)
			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				} else if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error to wrap %v, got %v", tt.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
