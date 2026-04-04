package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type authService interface {
	Register(ctx context.Context, email, password string, role model.Role) (uuid.UUID, error)
	Login(ctx context.Context, email, password string) (string, error)
	DummyLogin(context.Context, model.Role) (string, error)
}
