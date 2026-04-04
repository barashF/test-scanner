package schedule

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/jackc/pgx/v5"
)

const scheduleTable = "schedules"

type Repository struct {
	transactionManager manager
}

func NewRepository(transactionManager manager) *Repository {
	return &Repository{transactionManager: transactionManager}
}

func (repository *Repository) Create(ctx context.Context, schedule *model.Schedule) (uuid.UUID, error) {
	var id uuid.UUID
	connection, err := repository.transactionManager.GetConn(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
		INSERT INTO ` + scheduleTable + ` (id, room_id, days_of_week, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err = connection.QueryRow(ctx, query, schedule.ID, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("database error: %w", err)
	}
	return id, nil
}

func (repository *Repository) GetByRoomID(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error) {
	var schedule model.Schedule
	connection, err := repository.transactionManager.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
		SELECT id, room_id, days_of_week, start_time, end_time
		FROM ` + scheduleTable + `
		WHERE room_id = $1
	`

	err = connection.QueryRow(ctx, query, roomID).Scan(&schedule.ID, &schedule.RoomID, &schedule.DaysOfWeek, &schedule.StartTime, &schedule.EndTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &schedule, nil
}
