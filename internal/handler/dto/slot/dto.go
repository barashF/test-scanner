package slot

import (
	"time"

	"github.com/google/uuid"
)

type SlotResponse struct {
	ID       uuid.UUID `json:"id"`
	RoomID   uuid.UUID `json:"room_id"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	IsBooked bool      `json:"is_booked"`
}
