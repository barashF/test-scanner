package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/jackc/pgx/v5"
)

const bookingTable = "bookings"

type Repository struct {
	transactionManager manager
}

func NewRepository(transactionManager manager) *Repository {
	return &Repository{transactionManager: transactionManager}
}

func (r *Repository) Create(ctx context.Context, booking *model.Booking) (uuid.UUID, error) {
	var id uuid.UUID
	connection, err := r.transactionManager.GetConn(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
		INSERT INTO ` + bookingTable + ` (id, slot_id, user_id, status, conference_link, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err = connection.QueryRow(
		ctx,
		query,
		booking.ID,
		booking.SlotID,
		booking.UserID,
		booking.Status,
		booking.ConferenceLink,
		booking.CreatedAt,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("database error: %w", err)
	}
	return id, nil
}

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Booking, error) {
	connection, err := r.transactionManager.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
		SELECT booking.id, booking.slot_id, booking.user_id, booking.status, booking.conference_link, booking.created_at
		FROM ` + bookingTable + ` booking
		JOIN slots ON booking.slot_id = slots.id
		WHERE booking.user_id = $1 AND slots.start_time >= $2 AND status = $3
		ORDER BY slots.start_time ASC
	`

	rows, err := connection.Query(ctx, query, userID, time.Now().UTC(), model.BookingStatusActive)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var bookings []*model.Booking
	for rows.Next() {
		var booking model.Booking
		err := rows.Scan(
			&booking.ID,
			&booking.SlotID,
			&booking.UserID,
			&booking.Status,
			&booking.ConferenceLink,
			&booking.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		bookings = append(bookings, &booking)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if bookings == nil {
		bookings = []*model.Booking{}
	}

	return bookings, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, bookingID uuid.UUID, status model.BookingStatus) error {
	connection, err := r.transactionManager.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}

	query := `
		UPDATE ` + bookingTable + `
		SET status = $1
		WHERE id = $2
	`

	_, err = connection.Exec(ctx, query, status, bookingID)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	var booking model.Booking
	connection, err := r.transactionManager.GetConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	query := `
		SELECT id, slot_id, user_id, status, conference_link, created_at
		FROM ` + bookingTable + `
		WHERE id = $1
	`

	err = connection.QueryRow(ctx, query, id).Scan(
		&booking.ID,
		&booking.SlotID,
		&booking.UserID,
		&booking.Status,
		&booking.ConferenceLink,
		&booking.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &booking, nil
}

func (r *Repository) GetAllWithPagination(ctx context.Context, pageSize, offset int) ([]*model.Booking, int, error) {
	conn, err := r.transactionManager.GetConn(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("get conn: %w", err)
	}

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM bookings`
	err = conn.QueryRow(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, slot_id, user_id, status, conference_link, created_at
		FROM ` + bookingTable + `
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := conn.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var bookings []*model.Booking
	for rows.Next() {
		var booking model.Booking
		err := rows.Scan(
			&booking.ID,
			&booking.SlotID,
			&booking.UserID,
			&booking.Status,
			&booking.ConferenceLink,
			&booking.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan error: %w", err)
		}
		bookings = append(bookings, &booking)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("database error: %w", err)
	}

	if bookings == nil {
		bookings = []*model.Booking{}
	}

	return bookings, totalCount, nil
}
