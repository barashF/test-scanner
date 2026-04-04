package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/internships-backend/test-backend-barashF/internal/service/auth"
	"github.com/internships-backend/test-backend-barashF/internal/service/auth/mocks"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockuserRepository(ctrl)
	svc := auth.NewService(mockRepo, "test-secret-key-min-32-chars", time.Hour)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	role := model.UserRole

	t.Run("success", func(t *testing.T) {
		expectedID := uuid.New()
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(expectedID, nil).Times(1)

		id, err := svc.Register(ctx, email, password, role)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != expectedID {
			t.Errorf("expected id %v, got %v", expectedID, id)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(uuid.Nil, errors.New("db error")).Times(1)

		id, err := svc.Register(ctx, email, password, role)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if id != uuid.Nil {
			t.Errorf("expected nil id, got %v", id)
		}
	})
}

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockuserRepository(ctrl)
	svc := auth.NewService(mockRepo, "test-secret-key-min-32-chars", time.Hour)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	userID := uuid.New()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &model.User{
		ID:       userID,
		Email:    email,
		Password: string(hashedPassword),
		Role:     model.UserRole,
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().GetByEmail(ctx, email).Return(user, nil).Times(1)

		token, err := svc.Login(ctx, email, password)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Fatal("expected token, got empty string")
		}

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key-min-32-chars"), nil
		})
		if err != nil {
			t.Fatalf("failed to parse token: %v", err)
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("invalid token claims")
		}

		if claims["user_id"] != userID.String() {
			t.Errorf("expected user_id %v, got %v", userID.String(), claims["user_id"])
		}
		if claims["role"] != string(model.UserRole) {
			t.Errorf("expected role %v, got %v", model.UserRole, claims["role"])
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().GetByEmail(ctx, email).Return(nil, model.ErrNotFound).Times(1)

		token, err := svc.Login(ctx, email, password)

		if !errors.Is(err, model.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
		if token != "" {
			t.Errorf("expected empty token, got %v", token)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		mockRepo.EXPECT().GetByEmail(ctx, email).Return(user, nil).Times(1)

		token, err := svc.Login(ctx, email, "wrongpassword")

		if !errors.Is(err, model.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
		if token != "" {
			t.Errorf("expected empty token, got %v", token)
		}
	})
}

func TestService_DummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockuserRepository(ctrl)
	svc := auth.NewService(mockRepo, "test-secret-key-min-32-chars", time.Hour)

	ctx := context.Background()

	t.Run("admin role", func(t *testing.T) {
		token, err := svc.DummyLogin(ctx, model.AdminRole)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Fatal("expected token, got empty string")
		}

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key-min-32-chars"), nil
		})
		if err != nil {
			t.Fatalf("failed to parse token: %v", err)
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("invalid token claims")
		}

		if claims["role"] != string(model.AdminRole) {
			t.Errorf("expected role %v, got %v", model.AdminRole, claims["role"])
		}
	})

	t.Run("user role", func(t *testing.T) {
		token, err := svc.DummyLogin(ctx, model.UserRole)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Fatal("expected token, got empty string")
		}

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret-key-min-32-chars"), nil
		})
		if err != nil {
			t.Fatalf("failed to parse token: %v", err)
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("invalid token claims")
		}

		if claims["role"] != string(model.UserRole) {
			t.Errorf("expected role %v, got %v", model.UserRole, claims["role"])
		}
	})
}
