package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Role      Role
	CreatedAt time.Time
}

type Role string

var (
	UserRole  Role = "user"
	AdminRole Role = "admin"
)
