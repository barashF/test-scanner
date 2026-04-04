package validation

import (
	"strconv"

	"github.com/internships-backend/test-backend-barashF/internal/model"
)

func ValidatePage(pageStr string) error {
	if pageStr == "" {
		return nil
	}

	p, err := strconv.Atoi(pageStr)
	if err != nil || p < 1 {
		return model.ErrInvalidPagination
	}

	return nil
}

func ValidateSize(sizeStr string) error {
	if sizeStr == "" {
		return nil
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		return model.ErrInvalidPagination
	}

	return nil
}
