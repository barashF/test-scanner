package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	auth_controller "github.com/internships-backend/test-backend-barashF/internal/handler/auth"
	"github.com/internships-backend/test-backend-barashF/internal/handler/auth/mocks"
	dto "github.com/internships-backend/test-backend-barashF/internal/handler/dto/auth"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"go.uber.org/mock/gomock"
)

type noOpLogger struct{}

func (n noOpLogger) Debug(m string, f ...logger.Field) {}
func (n noOpLogger) Info(m string, f ...logger.Field)  {}
func (n noOpLogger) Warn(m string, f ...logger.Field)  {}
func (n noOpLogger) Error(m string, f ...logger.Field) {}
func (n noOpLogger) Sync() error                       { return nil }

func TestController_Register(t *testing.T) {
	type mockBehavior func(s *mocks.MockauthService)

	tests := []struct {
		name           string
		requestBody    any
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "user",
			},
			mockBehavior: func(s *mocks.MockauthService) {
				s.EXPECT().
					Register(gomock.Any(), "test@example.com", "password123", model.Role("user")).
					Return(uuid.New(), nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid-json",
			mockBehavior:   func(s *mocks.MockauthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Validation Failed (Empty Email)",
			requestBody: dto.RegisterRequest{
				Email:    "",
				Password: "password123",
				Role:     "user",
			},
			mockBehavior:   func(s *mocks.MockauthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error (Conflict/Email Exists)",
			requestBody: dto.RegisterRequest{
				Email:    "exists@example.com",
				Password: "password123",
				Role:     "user",
			},
			mockBehavior: func(s *mocks.MockauthService) {
				s.EXPECT().
					Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(uuid.Nil, model.ErrInvalidEmail)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockauthService(ctrl)
			tt.mockBehavior(mockService)

			handler := auth_controller.NewController(mockService, noOpLogger{})

			var buf bytes.Buffer
			if s, ok := tt.requestBody.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", &buf)
			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestController_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockauthService(ctrl)
	handler := auth_controller.NewController(mockService, noOpLogger{})

	t.Run("Success Login", func(t *testing.T) {
		email := "user@test.com"
		pass := "123456"
		token := "jwt-token-example"

		mockService.EXPECT().
			Login(gomock.Any(), email, pass).
			Return(token, nil)

		body, _ := json.Marshal(dto.LoginRequest{Email: email, Password: pass})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}

		var resp dto.AuthResponse
		json.NewDecoder(rr.Body).Decode(&resp)
		if resp.AccessToken != token {
			t.Errorf("expected token %s, got %s", token, resp.AccessToken)
		}
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		mockService.EXPECT().
			Login(gomock.Any(), gomock.Any(), gomock.Any()).
			Return("", model.ErrInvalidCredentials)

		body, _ := json.Marshal(dto.LoginRequest{Email: "wrong@test.com", Password: "wrong"})
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.Login(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rr.Code)
		}
	})
}

func TestController_DummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockauthService(ctrl)
	handler := auth_controller.NewController(mockService, noOpLogger{})

	t.Run("Success Dummy", func(t *testing.T) {
		token := "dummy-token"
		mockService.EXPECT().
			DummyLogin(gomock.Any(), model.Role("admin")).
			Return(token, nil)

		body, _ := json.Marshal(dto.DummyLoginRequest{Role: "admin"})
		req := httptest.NewRequest(http.MethodPost, "/dummy-login", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		handler.DummyLogin(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}
