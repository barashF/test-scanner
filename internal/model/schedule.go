package model

import "github.com/google/uuid"

type Schedule struct {
	ID         uuid.UUID
	RoomID     uuid.UUID
	DaysOfWeek DaysOfWeek
	StartTime  string
	EndTime    string
}

type DaysOfWeek []int
