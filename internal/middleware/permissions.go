package middleware

import (
	"net/http"

	"github.com/internships-backend/test-backend-barashF/internal/logger"
)

func RolesAllowed(log logger.Logger, allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userData, ok := r.Context().Value(UserContextKey).(UserData)

			if !ok {
				log.Warn("Access denied: no user data in context", logger.F("path", r.URL.Path))
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			isAllowed := false
			for _, role := range allowedRoles {
				if userData.Role == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				log.Warn("Access denied: insufficient permissions",
					logger.F("user_id", userData.UserID),
					logger.F("user_role", userData.Role),
					logger.F("required_roles", allowedRoles),
				)
				http.Error(w, `{"error":"Access denied: insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
