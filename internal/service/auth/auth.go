package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/internships-backend/test-backend-barashF/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo  userRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewService(repo userRepository, secret string, ttl time.Duration) *Service {
	return &Service{
		userRepo:  repo,
		jwtSecret: []byte(secret),
		tokenTTL:  ttl,
	}
}

func (s *Service) Register(ctx context.Context, email, password string, role model.Role) (uuid.UUID, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password error: %w", err)
	}

	newUser := &model.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}

	id, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create user error: %w", err)
	}

	return id, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", model.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", model.ErrInvalidCredentials
	}

	return s.generateJWT(user.ID, user.Role)
}

func (s *Service) DummyLogin(ctx context.Context, role model.Role) (string, error) {
	var userID uuid.UUID

	if role == model.AdminRole {
		userID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	} else {
		userID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	}

	return s.generateJWT(userID, role)
}

func (s *Service) generateJWT(userID uuid.UUID, role model.Role) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    string(role),
		"exp":     time.Now().Add(s.tokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token error: %w", err)
	}

	return signedToken, nil
}
