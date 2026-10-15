package categories

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

type CategoryService interface {
	ListCategories(ctx context.Context) ([]database.Category, error)
	GetCategoryByID(ctx context.Context, id int64) (database.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (database.Category, error)
	CreateCategory(ctx context.Context, categoryParams database.CreateCategoryParams) (database.Category, error)
	UpdateCategory(ctx context.Context, categoryParams database.UpdateCategoryParams) error
	DeleteCategory(ctx context.Context, id int64) error
}

type Handler struct {
	svc CategoryService
}

func NewHandler(svc CategoryService) *Handler {
	return &Handler{svc: svc}
}

// GET /categories
func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.ListCategories(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(categories)
}

// GET /categories/{id}
func (h *Handler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	categoryId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}

	category, err := h.svc.GetCategoryByID(r.Context(), categoryId)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(category)
}

// GET /categories/slug/{slug}
func (h *Handler) GetCategoryBySlug(w http.ResponseWriter, r *http.Request) {
	category, err := h.svc.GetCategoryBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(category)
}

// POST /categories
func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var categoryParams database.CreateCategoryParams

	err := json.NewDecoder(r.Body).Decode(&categoryParams)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := h.svc.CreateCategory(r.Context(), categoryParams)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// PUT /categories/{id}
func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}

	var categoryParams database.UpdateCategoryParams

	err = json.NewDecoder(r.Body).Decode(&categoryParams)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// id always comes from the url, never from the body
	categoryParams.ID = categoryId

	err = h.svc.UpdateCategory(r.Context(), categoryParams)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /categories/{id}
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid category id", http.StatusBadRequest)
		return
	}

	err = h.svc.DeleteCategory(r.Context(), categoryId)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeError maps domain errors to http status codes
func writeError(w http.ResponseWriter, err error) {
	log.Println(err)
	switch {
	case errors.Is(err, ErrNameRequired),
		errors.Is(err, ErrSlugRequired):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrSlugTaken):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, ErrCategoryNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
