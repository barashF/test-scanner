package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, getConnectString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "db unavailable: %v\n", err)
		os.Exit(1)
	}

	totalUsers := 10000
	totalRooms := 50
	totalSlots := 100000

	userIds := make([]uuid.UUID, totalUsers)
	roomIds := make([]uuid.UUID, totalRooms)
	slotIds := make([]uuid.UUID, totalSlots)

	userRows := make([][]any, 0, totalUsers)
	dummyHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

	for i := 0; i < totalUsers; i++ {
		id := uuid.New()
		userIds[i] = id
		userRows = append(userRows, []any{
			id,
			fmt.Sprintf("user%d@example.com", i),
			dummyHash,
			"user",
			time.Now(),
		})
	}

	_, err = pool.CopyFrom(
		ctx,
		pgx.Identifier{"users"},
		[]string{"id", "email", "password_hash", "role", "created_at"},
		pgx.CopyFromRows(userRows),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to copy users: %v\n", err)
		os.Exit(1)
	}

	roomRows := make([][]any, 0, totalRooms)
	for i := 0; i < totalRooms; i++ {
		id := uuid.New()
		roomIds[i] = id
		roomRows = append(roomRows, []any{
			id,
			fmt.Sprintf("Room %d", i+1),
			fmt.Sprintf("Description for room %d", i+1),
			10,
			time.Now(),
		})
	}

	_, err = pool.CopyFrom(
		ctx,
		pgx.Identifier{"rooms"},
		[]string{"id", "name", "description", "capacity", "created_at"},
		pgx.CopyFromRows(roomRows),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to copy rooms: %v\n", err)
		os.Exit(1)
	}

	slotRows := make([][]any, 0, totalSlots)
	startTime := time.Now().Truncate(time.Hour)

	for i := 0; i < totalSlots; i++ {
		id := uuid.New()
		slotIds[i] = id

		roomIdx := i % totalRooms
		seqInRoom := i / totalRooms

		slotStart := startTime.Add(time.Duration(seqInRoom*30) * time.Minute)
		slotEnd := slotStart.Add(30 * time.Minute)

		isBooked := i < totalSlots/2

		slotRows = append(slotRows, []any{
			id,
			roomIds[roomIdx],
			isBooked,
			slotStart,
			slotEnd,
		})
	}

	_, err = pool.CopyFrom(
		ctx,
		pgx.Identifier{"slots"},
		[]string{"id", "room_id", "is_booked", "start_time", "end_time"},
		pgx.CopyFromRows(slotRows),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to copy slots: %v\n", err)
		os.Exit(1)
	}

	totalBookings := totalSlots / 2
	bookingRows := make([][]any, 0, totalBookings)

	for i := 0; i < totalBookings; i++ {
		bookingRows = append(bookingRows, []any{
			uuid.New(),
			slotIds[i],
			userIds[i%totalUsers],
			"confirmed",
			fmt.Sprintf("https://telemost.yandex.ru/j/%d", i),
			time.Now(),
		})
	}

	_, err = pool.CopyFrom(
		ctx,
		pgx.Identifier{"bookings"},
		[]string{"id", "slot_id", "user_id", "status", "conference_link", "created_at"},
		pgx.CopyFromRows(bookingRows),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to copy bookings: %v\n", err)
		os.Exit(1)
	}
}

func getConnectString() string {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	db := os.Getenv("POSTGRES_DB")
	if db == "" {
		db = "booking_db"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, db)
}
