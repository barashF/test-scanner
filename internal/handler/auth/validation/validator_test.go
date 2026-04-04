package validation_test

import (
	"testing"

	"github.com/internships-backend/test-backend-barashF/internal/handler/auth/validation"
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/auth"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func TestValidateRequest_Register(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.RegisterRequest
		expectedErr error
	}{
		{
			name: "Success User",
			req: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "user",
			},
			expectedErr: nil,
		},
		{
			name: "Success Admin",
			req: &dto.RegisterRequest{
				Email:    "admin@example.com",
				Password: "password123",
				Role:     "admin",
			},
			expectedErr: nil,
		},
		{
			name: "Missing Email",
			req: &dto.RegisterRequest{
				Email:    "",
				Password: "password123",
				Role:     "user",
			},
			expectedErr: model.ErrMissingRequiredFields,
		},
		{
			name: "Missing Password",
			req: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "",
				Role:     "user",
			},
			expectedErr: model.ErrMissingRequiredFields,
		},
		{
			name: "Missing Role",
			req: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "",
			},
			expectedErr: model.ErrMissingRequiredFields,
		},
		{
			name: "Invalid Role",
			req: &dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "client",
			},
			expectedErr: model.ErrInvalidRole,
		},
		{
			name: "Invalid Email Format",
			req: &dto.RegisterRequest{
				Email:    "not-an-email",
				Password: "password123",
				Role:     "user",
			},
			expectedErr: model.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateRequest(tt.req)
			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidateRequest_Login(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.LoginRequest
		expectedErr error
	}{
		{
			name: "Success",
			req: &dto.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedErr: nil,
		},
		{
			name: "Missing Email",
			req: &dto.LoginRequest{
				Email:    "",
				Password: "password123",
			},
			expectedErr: model.ErrMissingRequiredFields,
		},
		{
			name: "Missing Password",
			req: &dto.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			},
			expectedErr: model.ErrMissingRequiredFields,
		},
		{
			name: "Invalid Email Format",
			req: &dto.LoginRequest{
				Email:    "test@",
				Password: "password123",
			},
			expectedErr: model.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateRequest(tt.req)
			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidateRequest_DummyLogin(t *testing.T) {
	tests := []struct {
		name        string
		req         *dto.DummyLoginRequest
		expectedErr error
	}{
		{
			name: "Success User",
			req: &dto.DummyLoginRequest{
				Role: "user",
			},
			expectedErr: nil,
		},
		{
			name: "Success Admin",
			req: &dto.DummyLoginRequest{
				Role: "admin",
			},
			expectedErr: nil,
		},
		{
			name: "Missing Role",
			req: &dto.DummyLoginRequest{
				Role: "",
			},
			expectedErr: model.ErrMissingRequiredFields,
		},
		{
			name: "Invalid Role",
			req: &dto.DummyLoginRequest{
				Role: "superadmin",
			},
			expectedErr: model.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateRequest(tt.req)
			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidateRequest_UnknownType(t *testing.T) {
	type unknownType struct{}
	req := &unknownType{}

	err := validation.ValidateRequest(req)
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}
