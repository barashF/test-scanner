package validation

import (
	"fmt"
	"net/mail"
	"strings"

	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/auth"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func ValidateRequest[T any](req *T) error {
	switch v := any(req).(type) {
	case *dto.RegisterRequest:
		return validateRegister(v)
	case *dto.LoginRequest:
		return validateLogin(v)
	case *dto.DummyLoginRequest:
		return validateDummyLogin(v)
	default:
		return fmt.Errorf("unknown type of request")
	}
}

func validateDummyLogin(req *dto.DummyLoginRequest) error {
	if req.Role == "" {
		return model.ErrMissingRequiredFields
	}

	err := validateRole(req.Role)
	if err != nil {
		return err
	}

	return nil
}

func validateLogin(req *dto.LoginRequest) error {
	if req.Email == "" || req.Password == "" {
		return model.ErrMissingRequiredFields
	}

	err := validateEmail(req.Email)
	if err != nil {
		return err
	}

	return nil
}

func validateRegister(req *dto.RegisterRequest) error {
	if req.Email == "" || req.Password == "" || req.Role == "" {
		return model.ErrMissingRequiredFields
	}

	err := validateEmail(req.Email)
	if err != nil {
		return err
	}

	err = validateRole(req.Role)
	if err != nil {
		return err
	}

	return nil
}

func validateRole(role string) error {
	switch role {
	case string(model.UserRole), string(model.AdminRole):
		return nil
	default:
		return model.ErrInvalidRole
	}
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)

	if len(email) < 5 || len(email) > 254 {
		return model.ErrInvalidEmail
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("%w: %v", model.ErrInvalidEmail, err)
	}

	return nil
}
