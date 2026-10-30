package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	database "github.com/bekasyldev/lumea/internal/database"
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var u database.CreateUserParams

	err := json.NewDecoder(r.Body).Decode((&u))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createUser, err := h.svc.Register(r.Context(), u.Username, u.HashedPassword)
	if err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createUser)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var u database.CreateUserParams

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokenPair, err := h.svc.Login(r.Context(), u.Username, u.HashedPassword)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token.TokenPair{AccessToken: tokenPair.AccessToken})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token.TokenPair{AccessToken: accessToken})
}
