package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type userRepository interface {
	Create(ctx context.Context, user *model.User) (uuid.UUID, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}
