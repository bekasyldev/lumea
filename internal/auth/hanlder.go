package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/bekasyldev/lumea/internal/json"
	"github.com/bekasyldev/lumea/internal/token"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (database.CreateUserRow, error)
	Login(ctx context.Context, username, password string) (token.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (string, error)
}

type AuthHandler struct {
	svc AuthService
}

func NewAuthHandler(svc AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var c credentials
	if err := json.Read(r, &c); err != nil {
		slog.WarnContext(ctx, "register: invalid request body", "err", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	createUser, err := h.svc.Register(ctx, c.Username, c.Password)
	if err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			slog.WarnContext(ctx, "register: username taken", "username", c.Username)
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		slog.ErrorContext(ctx, "register: create user failed", "err", err, "username", c.Username)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusCreated, createUser)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var c credentials
	if err := json.Read(r, &c); err != nil {
		slog.WarnContext(ctx, "login: invalid request body", "err", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tokenPair, err := h.svc.Login(ctx, c.Username, c.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			slog.WarnContext(ctx, "login: invalid credentials", "err", err, "username", c.Username)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		slog.ErrorContext(ctx, "login: failed", "err", err, "username", c.Username)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// store refresh token in an httpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60,
		Path:     "/auth/refresh",
	})

	json.Write(w, http.StatusOK, token.TokenPair{AccessToken: tokenPair.AccessToken})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		slog.WarnContext(ctx, "token: missing refresh token", "err", err)
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.svc.Refresh(ctx, cookie.Value)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			slog.WarnContext(ctx, "token: invalid refresh token", "err", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		slog.ErrorContext(ctx, "token: refresh failed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusOK, token.TokenPair{AccessToken: accessToken})
}
