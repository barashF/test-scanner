package booking

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type Service struct {
	bookingRepo       bookingRepository
	slotRepo          slotRepository
	conferenseGateway conferenceGateway
	manager           manager
	logger            logger.Logger
}

func NewService(
	bookingRepo bookingRepository,
	slotRepo slotRepository,
	conf conferenceGateway,
	manager manager,
	logger logger.Logger,
) *Service {
	return &Service{
		bookingRepo:       bookingRepo,
		slotRepo:          slotRepo,
		conferenseGateway: conf,
		manager:           manager,
		logger:            logger,
	}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, slotID uuid.UUID, createConferenceLink bool) (*model.Booking, error) {
	slot, err := s.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return nil, err
	}

	if slot.Start.Before(time.Now().UTC()) {
		return nil, model.ErrInvalidTime
	}

	if slot.IsBooked {
		return nil, model.ErrIsBooking
	}

	booking := &model.Booking{
		ID:        uuid.New(),
		SlotID:    slotID,
		UserID:    userID,
		Status:    model.BookingStatusActive,
		CreatedAt: time.Now().UTC(),
	}

	if createConferenceLink {
		link, err := s.conferenseGateway.CreateConference(ctx, userID, slot.Start)
		if err != nil {
			s.logger.Warn("Conference service not response",
				logger.F("error", err.Error()),
			)
		}
		booking.ConferenceLink = link
	}

	err = s.manager.InTransaction(ctx, nil, func(dbContext context.Context) error {
		_, err := s.bookingRepo.Create(dbContext, booking)
		if err != nil {
			return err
		}

		err = s.slotRepo.UpdateIsBooked(dbContext, slotID, true)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]*model.Booking, int, error) {
	offset := (page - 1) * pageSize

	bookings, totalCount, err := s.bookingRepo.GetAllWithPagination(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return bookings, totalCount, nil
}

func (s *Service) GetByUser(ctx context.Context, userID uuid.UUID) ([]*model.Booking, error) {
	bookings, err := s.bookingRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (s *Service) Cancel(ctx context.Context, bookingID, userID uuid.UUID) (*model.Booking, error) {
	var updatedBooking *model.Booking

	err := s.manager.InTransaction(ctx, nil, func(dbContext context.Context) error {
		booking, err := s.bookingRepo.GetByID(dbContext, bookingID)
		if err != nil {
			return err
		}

		if booking.UserID != userID {
			return model.ErrBookingOwnershipRequired
		}

		if booking.Status == model.BookingStatusCancelled {
			updatedBooking = booking
			return nil
		}

		err = s.bookingRepo.UpdateStatus(dbContext, bookingID, model.BookingStatusCancelled)
		if err != nil {
			return err
		}

		err = s.slotRepo.UpdateIsBooked(dbContext, booking.SlotID, false)
		if err != nil {
			return err
		}

		booking.Status = model.BookingStatusCancelled
		updatedBooking = booking

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedBooking, nil
}
