package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/bekasyldev/lumea/internal/repository"
	"github.com/go-chi/chi/v5"
)

type UserService interface {
	GetById(ctx context.Context, id int64) (database.GetUserRow, error)
	GetList(ctx context.Context) ([]database.ListUsersRow, error)
}

type UserHandler struct {
	svc UserService
}

func NewUserHanlder(svc UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) GetById(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.svc.GetById(r.Context(), id)
	if err != nil {
		log.Println(err)
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
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

func (h *UserHandler) GetList(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.GetList(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
