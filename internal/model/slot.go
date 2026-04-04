package model

import (
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	ID       uuid.UUID
	RoomID   uuid.UUID
	IsBooked bool
	Start    time.Time
	End      time.Time
}
