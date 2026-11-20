package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/bekasyldev/lumea/internal/token"
)

type contextKey string

const userContextKey contextKey = "user"

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or malformed token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &token.Claims{}

		token, err := token.ValidateToken(tokenString, token.TokenAccess, claims)

		if err != nil || !token.Valid {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) requireRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := r.Context().Value(userContextKey)
		claims, ok := c.(*token.Claims)
		if !ok {
			http.Error(w, "role is not provided", http.StatusBadRequest)
			return
		}
		if claims.Role != "admin" {
			http.Error(w, "not have access", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
