package user

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/go-chi/chi/v5"
)

type userService interface {
	GetByID(ctx context.Context, id int64) (database.GetUserRow, error)
	List(ctx context.Context) ([]database.ListUsersRow, error)
}

type Handler struct {
	svc userService
}

func NewHandler(svc userService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		log.Println(err)
		switch {
		case errors.Is(err, ErrUserNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
