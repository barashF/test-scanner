package validation_test

import (
	"testing"

	"github.com/internships-backend/test-backend-barashF/internal/handler/booking/validation"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func TestValidatePage(t *testing.T) {
	tests := []struct {
		name        string
		pageStr     string
		expectedErr error
	}{
		{
			name:        "Empty string (defaults to 1, valid)",
			pageStr:     "",
			expectedErr: nil,
		},
		{
			name:        "Valid page number",
			pageStr:     "5",
			expectedErr: nil,
		},
		{
			name:        "Boundary value (lowest valid)",
			pageStr:     "1",
			expectedErr: nil,
		},
		{
			name:        "Zero page (invalid)",
			pageStr:     "0",
			expectedErr: model.ErrInvalidPagination,
		},
		{
			name:        "Negative page (invalid)",
			pageStr:     "-5",
			expectedErr: model.ErrInvalidPagination,
		},
		{
			name:        "Not a number",
			pageStr:     "abc",
			expectedErr: model.ErrInvalidPagination,
		},
		{
			name:        "Float string",
			pageStr:     "1.5",
			expectedErr: model.ErrInvalidPagination,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidatePage(tt.pageStr)
			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestValidateSize(t *testing.T) {
	tests := []struct {
		name        string
		sizeStr     string
		expectedErr error
	}{
		{
			name:        "Empty string (defaults to 20, valid)",
			sizeStr:     "",
			expectedErr: nil,
		},
		{
			name:        "Valid size within range",
			sizeStr:     "50",
			expectedErr: nil,
		},
		{
			name:        "Boundary value (lowest valid)",
			sizeStr:     "1",
			expectedErr: nil,
		},
		{
			name:        "Boundary value (highest valid)",
			sizeStr:     "100",
			expectedErr: nil,
		},
		{
			name:        "Zero size (invalid)",
			sizeStr:     "0",
			expectedErr: model.ErrInvalidPagination,
		},
		{
			name:        "Negative size (invalid)",
			sizeStr:     "-10",
			expectedErr: model.ErrInvalidPagination,
		},
		{
			name:        "Size too large (invalid)",
			sizeStr:     "101",
			expectedErr: model.ErrInvalidPagination,
		},
		{
			name:        "Not a number",
			sizeStr:     "twenty",
			expectedErr: model.ErrInvalidPagination,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateSize(tt.sizeStr)
			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
