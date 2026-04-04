package booking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"

	// Правильные импорты сервиса и локальных моков
	"github.com/internships-backend/test-backend-barashF/internal/service/booking"
	"github.com/internships-backend/test-backend-barashF/internal/service/booking/mocks"

	"go.uber.org/mock/gomock"
)

func TestCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBookingRepo := mocks.NewMockbookingRepository(ctrl)
	mockSlotRepo := mocks.NewMockslotRepository(ctrl)
	mockConfGateway := mocks.NewMockconferenceGateway(ctrl)
	mockManager := mocks.NewMockmanager(ctrl)

	s := booking.NewService(mockBookingRepo, mockSlotRepo, mockConfGateway, mockManager, nil)

	userID := uuid.New()
	slotID := uuid.New()
	slot := &model.Slot{
		ID:       slotID,
		Start:    time.Now().Add(24 * time.Hour).UTC(),
		IsBooked: false,
	}

	mockSlotRepo.EXPECT().
		GetByID(gomock.Any(), slotID).
		Return(slot, nil).
		Times(1)

	mockConfGateway.EXPECT().
		CreateConference(gomock.Any(), userID, slot.Start).
		Return("http://conference.link", nil).
		Times(1)

	mockManager.EXPECT().
		InTransaction(gomock.Any(), gomock.Nil(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts *model.TransactionOptions, fn func(context.Context) error) error {
			return fn(ctx)
		}).Times(1)

	mockBookingRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(uuid.New(), nil).
		Times(1)

	mockSlotRepo.EXPECT().
		UpdateIsBooked(gomock.Any(), slotID, true).
		Return(nil).
		Times(1)

	b, err := s.Create(context.Background(), userID, slotID, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if b.ConferenceLink != "http://conference.link" {
		t.Errorf("expected link http://conference.link, got %s", b.ConferenceLink)
	}
}

func TestCreate_SlotAlreadyBooked(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSlotRepo := mocks.NewMockslotRepository(ctrl)
	mockManager := mocks.NewMockmanager(ctrl)

	s := booking.NewService(nil, mockSlotRepo, nil, mockManager, nil)

	slotID := uuid.New()
	slot := &model.Slot{
		ID:       slotID,
		Start:    time.Now().Add(24 * time.Hour).UTC(),
		IsBooked: true,
	}

	mockSlotRepo.EXPECT().
		GetByID(gomock.Any(), slotID).
		Return(slot, nil).
		Times(1)

	_, err := s.Create(context.Background(), uuid.New(), slotID, false)

	if !errors.Is(err, model.ErrIsBooking) {
		t.Errorf("expected ErrIsBooking, got %v", err)
	}
}

func TestCancel_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBookingRepo := mocks.NewMockbookingRepository(ctrl)
	mockSlotRepo := mocks.NewMockslotRepository(ctrl)
	mockManager := mocks.NewMockmanager(ctrl)

	s := booking.NewService(mockBookingRepo, mockSlotRepo, nil, mockManager, nil)

	bookingID := uuid.New()
	userID := uuid.New()
	slotID := uuid.New()
	b := &model.Booking{
		ID:     bookingID,
		UserID: userID,
		SlotID: slotID,
		Status: model.BookingStatusActive,
	}

	mockManager.EXPECT().
		InTransaction(gomock.Any(), gomock.Nil(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts *model.TransactionOptions, fn func(context.Context) error) error {
			return fn(ctx)
		}).Times(1)

	mockBookingRepo.EXPECT().
		GetByID(gomock.Any(), bookingID).
		Return(b, nil).
		Times(1)

	mockBookingRepo.EXPECT().
		UpdateStatus(gomock.Any(), bookingID, model.BookingStatusCancelled).
		Return(nil).
		Times(1)

	mockSlotRepo.EXPECT().
		UpdateIsBooked(gomock.Any(), slotID, false).
		Return(nil).
		Times(1)

	updatedBooking, err := s.Cancel(context.Background(), bookingID, userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updatedBooking.Status != model.BookingStatusCancelled {
		t.Errorf("expected status cancelled, got %s", updatedBooking.Status)
	}
}

func TestCancel_NotOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBookingRepo := mocks.NewMockbookingRepository(ctrl)
	mockManager := mocks.NewMockmanager(ctrl)

	s := booking.NewService(mockBookingRepo, nil, nil, mockManager, nil)

	bookingID := uuid.New()
	b := &model.Booking{
		ID:     bookingID,
		UserID: uuid.New(), // другой юзер
		Status: model.BookingStatusActive,
	}

	mockManager.EXPECT().
		InTransaction(gomock.Any(), gomock.Nil(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts *model.TransactionOptions, fn func(context.Context) error) error {
			return fn(ctx)
		}).Times(1)

	mockBookingRepo.EXPECT().
		GetByID(gomock.Any(), bookingID).
		Return(b, nil).
		Times(1)

	_, err := s.Cancel(context.Background(), bookingID, uuid.New())

	if !errors.Is(err, model.ErrBookingOwnershipRequired) {
		t.Errorf("expected ErrBookingOwnershipRequired, got %v", err)
	}
}

func TestList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockBookingRepo := mocks.NewMockbookingRepository(ctrl)

	s := booking.NewService(mockBookingRepo, nil, nil, nil, nil)

	mockBookings := []*model.Booking{{ID: uuid.New()}}

	mockBookingRepo.EXPECT().
		GetAllWithPagination(gomock.Any(), 10, 20).
		Return(mockBookings, 1, nil).
		Times(1)

	bookings, total, err := s.List(context.Background(), 3, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(bookings) != 1 {
		t.Errorf("expected 1 booking, got %d", len(bookings))
	}
}
