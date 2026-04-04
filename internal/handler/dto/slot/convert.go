package slot

import "github.com/internships-backend/test-backend-barashF/internal/model"

func ModelToResponse(slot *model.Slot) SlotResponse {
	return SlotResponse{
		ID:       slot.ID,
		RoomID:   slot.RoomID,
		Start:    slot.Start,
		End:      slot.End,
		IsBooked: slot.IsBooked,
	}
}
