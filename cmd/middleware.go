package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/bekasyldev/lumea/internal/token"
)

type contextKey string

const userContextKey contextKey = "user"

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or malformed token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &token.Claims{}

		token, err := token.ValidateToken(tokenString, claims)

		if err != nil || !token.Valid {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
