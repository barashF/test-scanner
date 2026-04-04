package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/jackc/pgx/v5"
)

const userTable = "users"

type Repository struct {
	transactionManager manager
}

func NewRepository(transactionManager manager) *Repository {
	return &Repository{transactionManager: transactionManager}
}

func (repository *Repository) Create(ctx context.Context, user *model.User) (uuid.UUID, error) {
	connection, err := repository.transactionManager.GetConn(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
		INSERT INTO ` + userTable + ` (id, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id uuid.UUID
	err = connection.QueryRow(ctx, query, user.ID, user.Email, user.Password, user.Role, user.CreatedAt).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("database error: %w", err)
	}
	return id, nil
}

func (repository *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	connection, err := repository.transactionManager.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
		SELECT id, email, password_hash, role, created_at
		FROM ` + userTable + `
		WHERE id = $1
	`

	err = connection.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}

func (repository *Repository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	connection, err := repository.transactionManager.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
		SELECT id, email, password_hash, role, created_at
		FROM ` + userTable + `
		WHERE email = $1
	`

	err = connection.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}
