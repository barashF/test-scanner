package room

import "github.com/google/uuid"

type CreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Capacity    int    `json:"capacity,omitempty"`
}

type RoomResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Capacity    int       `json:"capacity"`
}
