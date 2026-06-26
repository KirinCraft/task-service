package middleware

import (
	"context"
	"net/http"
	"strings"

	"task-service/internal/auth"
)

type contextKey string

const (
	userIDKey    contextKey = "user_id"
	bearerPrefix string     = "Bearer "
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

func NewAuth(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString, ok := strings.CutPrefix(authHeader, bearerPrefix)
		tokenString = strings.TrimSpace(tokenString)

		if !ok || tokenString == "" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		userID, err := m.jwtManager.Parse(tokenString)

		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}