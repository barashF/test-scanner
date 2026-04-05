package auth

import "github.com/google/uuid"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DummyLoginRequest struct {
	Role string `json:"role"`
}

type RegisterResponse struct {
	ID      uuid.UUID `json:"id"`
	Message string    `json:"message"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"текст ошибки или причина отказа"`
}
