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

type SlotListResponse struct {
	Slots []SlotResponse `json:"slots"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"текст ошибки или причина отказа"`
}
