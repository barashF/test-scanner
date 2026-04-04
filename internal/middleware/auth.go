package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/internships-backend/test-backend-barashF/internal/logger"
)

type contextKey string

const UserContextKey contextKey = "user_data"

type UserData struct {
	UserID string
	Role   string
}

func AuthMiddleware(jwtSecret string, log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn("Missing authorization header", logger.F("path", r.URL.Path))
				http.Error(w, `{"error":"Missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Warn("Invalid authorization header format", logger.F("path", r.URL.Path))
				http.Error(w, `{"error":"Invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				log.Warn("Invalid or expired token", logger.F("error", err.Error()))
				http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"Failed to parse token claims"}`, http.StatusUnauthorized)
				return
			}

			userID, _ := claims["user_id"].(string)
			role, _ := claims["role"].(string)

			userData := UserData{UserID: userID, Role: role}
			ctx := context.WithValue(r.Context(), UserContextKey, userData)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
