package room

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/jackc/pgx/v5"
)

const roomTable = "rooms"

type Repository struct {
	transactionManager manager
}

func NewRepository(transactionManager manager) *Repository {
	return &Repository{
		transactionManager: transactionManager,
	}
}

func (r *Repository) Create(ctx context.Context, room *model.Room) (uuid.UUID, error) {
	var id uuid.UUID

	conn, err := r.transactionManager.GetConn(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get conn from transaction manager: %w", err)
	}

	query := `
		INSERT INTO ` + roomTable + fmt.Sprintf(` (id, %s, description, capacity, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, room.Name)

	err = conn.QueryRow(ctx, query,
		room.ID,
		room.Name,
		room.Description,
		room.Capacity,
		room.CreatedAt,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("database error: %w", err)
	}

	return id, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	var room model.Room

	conn, err := r.transactionManager.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get conn from transaction manager: %w", err)
	}

	query := `
		SELECT id, name, description, capacity, created_at
		FROM ` + roomTable + `
		WHERE id = $1
	`

	err = conn.QueryRow(ctx, query, id).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.Capacity,
		&room.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &room, nil
}

func (r *Repository) GetAll(ctx context.Context) ([]*model.Room, error) {
	conn, err := r.transactionManager.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
        SELECT id, name, description, capacity, created_at
        FROM ` + roomTable + `
        ORDER BY created_at DESC
    `

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	var rooms []*model.Room
	for rows.Next() {
		var room model.Room
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.Capacity,
			&room.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		rooms = append(rooms, &room)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return rooms, nil
}
