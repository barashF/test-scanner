package validation_test

import (
	"errors"
	"strings"
	"testing"

	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/room"
	"github.com/internships-backend/test-backend-barashF/internal/handler/room/validation"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func TestValidateCreateRequestRoom(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.CreateRequest
		expectedErr error
	}{
		{
			name: "Success",
			req: &dto.CreateRequest{
				Name:        "Conference Room A",
				Description: "Equipped with projector",
				Capacity:    10,
			},
			expectedErr: nil,
		},
		{
			name: "Success with Empty Description",
			req: &dto.CreateRequest{
				Name:        "Conference Room A",
				Description: "",
				Capacity:    10,
			},
			expectedErr: nil,
		},
		{
			name: "Success with Zero Capacity",
			req: &dto.CreateRequest{
				Name:        "Virtual Room",
				Description: "Online only",
				Capacity:    0,
			},
			expectedErr: nil,
		},
		{
			name: "Empty Name",
			req: &dto.CreateRequest{
				Name:        "",
				Description: "Description",
				Capacity:    10,
			},
			expectedErr: model.ErrMissingRequiredFields,
		},
		{
			name: "Spaces Only Name",
			req: &dto.CreateRequest{
				Name:        "   ",
				Description: "Description",
				Capacity:    10,
			},
			expectedErr: model.ErrMissingRequiredFields,
		},
		{
			name: "Name Too Long",
			req: &dto.CreateRequest{
				Name:        strings.Repeat("a", 256),
				Description: "Description",
				Capacity:    10,
			},
			expectedErr: model.ErrValidationFailed,
		},
		{
			name: "Negative Capacity",
			req: &dto.CreateRequest{
				Name:        "Conference Room A",
				Description: "Description",
				Capacity:    -1,
			},
			expectedErr: model.ErrValidationFailed,
		},
		{
			name: "Capacity Too Large",
			req: &dto.CreateRequest{
				Name:        "Stadium",
				Description: "Too many people",
				Capacity:    1001,
			},
			expectedErr: model.ErrValidationFailed,
		},
		{
			name: "Description Too Long",
			req: &dto.CreateRequest{
				Name:        "Conference Room A",
				Description: strings.Repeat("a", 1001),
				Capacity:    10,
			},
			expectedErr: model.ErrValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateCreateRequestRoom(tt.req)
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
