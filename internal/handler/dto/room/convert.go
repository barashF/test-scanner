package room

import "github.com/internships-backend/test-backend-barashF/internal/model"

func ModelToResponse(room *model.Room) RoomResponse {
	return RoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		Capacity:    room.Capacity,
	}
}
