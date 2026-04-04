package validation

import (
	"fmt"
	"strings"

	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/room"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func ValidateCreateRequestRoom(req *dto.CreateRequest) error {
	err := validateRoomName(req.Name)
	if err != nil {
		return err
	}

	err = validateRoomDescription(req.Description)
	if err != nil {
		return err
	}

	err = validateRoomCapacity(req.Capacity)
	if err != nil {
		return err
	}

	return nil
}

func validateRoomName(name string) error {
	trimmed := strings.TrimSpace(name)

	if trimmed == "" {
		return fmt.Errorf("room name cannot be empty or only spaces: %w", model.ErrMissingRequiredFields)
	}

	if len(trimmed) > 255 {
		return fmt.Errorf("room name is too long (max 255 chars): %w", model.ErrValidationFailed)
	}

	return nil
}

func validateRoomCapacity(capacity int) error {
	if capacity < 0 {
		return fmt.Errorf("capacity must be a positive number: %w", model.ErrValidationFailed)
	}

	if capacity > 1000 {
		return fmt.Errorf("capacity cannot exceed 1000: %w", model.ErrValidationFailed)
	}

	return nil
}

func validateRoomDescription(description string) error {
	if len(description) > 1000 {
		return fmt.Errorf("description is too long (max 1000 chars): %w", model.ErrValidationFailed)
	}

	return nil
}
