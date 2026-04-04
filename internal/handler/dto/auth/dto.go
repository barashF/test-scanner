package auth

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

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}
