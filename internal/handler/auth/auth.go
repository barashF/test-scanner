package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/internships-backend/test-backend-barashF/internal/handler/auth/validation"
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/auth"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
)

type Controller struct {
	authService authService
	logger      logger.Logger
}

func NewController(service authService, logger logger.Logger) *Controller {
	return &Controller{
		authService: service,
		logger:      logger,
	}
}

func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Warn("Failed to decode register request body", logger.F("error", err.Error()))

		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	err := validation.ValidateRequest(&req)
	if err != nil {
		c.logger.Warn("Register request validation failed",
			logger.F("email", req.Email),
			logger.F("error", err.Error()),
		)

		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	id, err := c.authService.Register(r.Context(), req.Email, req.Password, model.Role(req.Role))
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	c.logger.Info("User registered successfully", logger.F("user_id", id))

	c.writeJSON(w, http.StatusCreated, map[string]any{
		"id":      id,
		"message": "User registered successfully",
	})
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Warn("Failed to decode login request body", logger.F("error", err.Error()))

		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	token, err := c.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	c.writeJSON(w, http.StatusOK, dto.AuthResponse{AccessToken: token})
}

func (c *Controller) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.DummyLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Warn("Failed to decode dummy login request body", logger.F("error", err.Error()))

		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	err := validation.ValidateRequest(&req)
	if err != nil {
		c.logger.Warn("Dummy login request validation failed", logger.F("error", err.Error()))

		c.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	token, err := c.authService.DummyLogin(r.Context(), model.Role(req.Role))
	if err != nil {
		c.writeError(w, err, r)
		return
	}

	c.writeJSON(w, http.StatusOK, dto.AuthResponse{AccessToken: token})
}

func (c *Controller) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		c.logger.Error("Failed to encode response JSON", logger.F("error", err.Error()))

		http.Error(w, `{"error":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}

func (c *Controller) writeError(w http.ResponseWriter, err error, r *http.Request) {
	message, status := c.mapError(err)
	if status == http.StatusInternalServerError {
		c.logger.Error("Internal server error occurred",
			logger.F("error", err.Error()),
			logger.F("path", r.URL.Path),
		)
	} else {
		c.logger.Warn("Request failed with client error",
			logger.F("status", status),
			logger.F("error", err.Error()),
			logger.F("path", r.URL.Path),
		)
	}

	c.writeJSON(w, status, map[string]string{"error": message})
}

func (c *Controller) mapError(err error) (string, int) {
	switch {
	case errors.Is(err, model.ErrInvalidCredentials):
		return "Invalid email or password", http.StatusUnauthorized

	case errors.Is(err, model.ErrMissingRequiredFields),
		errors.Is(err, model.ErrInvalidRole),
		errors.Is(err, model.ErrInvalidEmail):
		return err.Error(), http.StatusBadRequest

	default:
		return "Internal server error", http.StatusInternalServerError
	}
}
