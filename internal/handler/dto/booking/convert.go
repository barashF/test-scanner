package booking

import "github.com/internships-backend/test-backend-barashF/internal/model"

func ModelToResponse(booking *model.Booking) Response {
	return Response{
		ID:             booking.ID,
		SlotID:         booking.SlotID,
		UserID:         booking.UserID,
		Status:         string(booking.Status),
		ConferenceLink: booking.ConferenceLink,
		CreatedAt:      booking.CreatedAt,
	}
}
