package booking

import (
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	SlotID               uuid.UUID `json:"slot_id"`
	CreateConferenceLink bool      `json:"create_conference_link"`
}

type Response struct {
	ID             uuid.UUID `json:"id"`
	SlotID         uuid.UUID `json:"slot_id"`
	UserID         uuid.UUID `json:"user_id"`
	Status         string    `json:"status"`
	ConferenceLink string    `json:"conference_link,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
