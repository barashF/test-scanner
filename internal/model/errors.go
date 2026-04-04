package model

import "errors"

var (
	ErrMissingRequiredFields = errors.New("missing required fields")
	ErrInvalidRole           = errors.New("invalid role")
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidCredentials    = errors.New("invalid email or password")
)

var (
	ErrScheduleExists    = errors.New("schedule for this room already exists and cannot be changed")
	ErrInvalidTime       = errors.New("invalid time")
	ErrInvalidDaysOfWeek = errors.New("invalid day of week")
	ErrNotFound          = errors.New("not found")
)

var (
	ErrValidationFailed         = errors.New("provided data failed validation")
	ErrIsBooking                = errors.New("slot already is booking")
	ErrBookingOwnershipRequired = errors.New("cannot cancel another user's booking")
	ErrInvalidPagination        = errors.New("invalid pagination")
)
