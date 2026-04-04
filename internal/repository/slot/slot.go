package slot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/jackc/pgx/v5"
)

const slotTable = "slots"

type Repository struct {
	manager manager
}

func NewRepository(manager manager) *Repository {
	return &Repository{manager: manager}
}

func (r *Repository) GetByRoomAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]*model.Slot, error) {
	conn, err := r.manager.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT id, room_id, is_booked, start_time, end_time
		FROM ` + slotTable + `
		WHERE room_id = $1 AND start_time >= $2 AND start_time < $3 AND is_booked = false
		ORDER BY start_time ASC
	`

	rows, err := conn.Query(ctx, query, roomID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var slots []*model.Slot
	for rows.Next() {
		var slot model.Slot
		if err := rows.Scan(&slot.ID, &slot.RoomID, &slot.IsBooked, &slot.Start, &slot.End); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		slots = append(slots, &slot)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if slots == nil {
		slots = []*model.Slot{}
	}
	return slots, nil
}

func (r *Repository) CreateMany(ctx context.Context, slots []*model.Slot) error {
	if len(slots) == 0 {
		return nil
	}
	conn, err := r.manager.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}

	query := `INSERT INTO ` + slotTable + ` (id, room_id, start_time, end_time) VALUES `
	arguments := []interface{}{}

	for index, slot := range slots {
		position := index * 4
		query += fmt.Sprintf("($%d, $%d, $%d, $%d),", position+1, position+2, position+3, position+4)
		arguments = append(arguments, slot.ID, slot.RoomID, slot.Start, slot.End)
	}
	query = query[:len(query)-1]

	_, err = conn.Exec(ctx, query, arguments...)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Slot, error) {
	conn, err := r.manager.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
		SELECT id, room_id, is_booked, start_time, end_time
		FROM ` + slotTable + `
		WHERE id = $1
	`

	var slot model.Slot
	err = conn.QueryRow(ctx, query, id).Scan(&slot.ID, &slot.RoomID, &slot.IsBooked, &slot.Start, &slot.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &slot, nil
}

func (r *Repository) UpdateIsBooked(ctx context.Context, slotID uuid.UUID, isBooked bool) error {
	conn, err := r.manager.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}

	query := `
		UPDATE ` + slotTable + `
		SET is_booked = $1
		WHERE id = $2
	`

	result, err := conn.Exec(ctx, query, isBooked, slotID)
	if err != nil {
		return fmt.Errorf("failed to update slot status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return model.ErrNotFound
	}

	return nil
}
