package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type UserSeed struct {
	ID       uuid.UUID
	Email    string
	Password string
	Role     string
}

func SeedUsers(ctx context.Context, db DB, logger logger.Logger) error {
	users := []UserSeed{
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Email:    "admin@gmail.com",
			Password: "avito.tech",
			Role:     "admin",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Email:    "user@gmail.com",
			Password: "avito.tech",
			Role:     "user",
		},
	}

	query := `
		INSERT INTO users (id, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`

	logger.Info("Starting database seeding for users...")

	count := 0
	for _, u := range users {
		_, err := db.Exec(ctx, query, u.ID, u.Email, u.Password, u.Role, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("failed to seed user %s: %w", u.ID, err)
		}
		count++
	}

	logger.Info("Database seeding completed")
	return nil
}
