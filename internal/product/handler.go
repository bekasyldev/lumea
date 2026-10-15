package product

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

type ProductService interface {
	ListProducts(ctx context.Context) ([]database.Product, error)
	CreateProduct(ctx context.Context, productParams database.CreateProductParams) error
	ListProductVariants(ctx context.Context, productId int64) ([]database.ProductVariant, error)
	CreateProductVariant(ctx context.Context, variantParams database.CreateProductVariantParams) error
}

type Handler struct {
	svc ProductService
}

func NewHandler(svc ProductService) *Handler {
	return &Handler{svc: svc}
}

// GET /products
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.svc.ListProducts(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

// POST /products
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product database.CreateProductParams

	parseErr := json.NewDecoder(r.Body).Decode((&product))
	if parseErr != nil {
		log.Println(parseErr)
		http.Error(w, parseErr.Error(), http.StatusBadRequest)
		return
	}

	err := h.svc.CreateProduct(r.Context(), product)
	if err != nil {
		switch {
		case errors.Is(err, ErrNameRequired) || errors.Is(err, ErrSlugRequired):
			log.Println(err)
			http.Error(w, "invalid credentials", http.StatusBadRequest)
			return
		case errors.Is(err, ErrSlugTaken):
			log.Println(err)
			http.Error(w, "slug alreday in use", http.StatusBadRequest)
			return
		case errors.Is(err, ErrCategoryNotFound):
			log.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

}

// GET /products/{id}/variants
func (h *Handler) ListProductVariants(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	productId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	productVariants, err := h.svc.ListProductVariants(r.Context(), productId)
	if err != nil {
		log.Println(err)
		switch {
		case errors.Is(err, ErrProductNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(productVariants)
}

// POST /products/{id}/variants
func (h *Handler) CreateProductVariant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	var productVariant database.CreateProductVariantParams

	err = json.NewDecoder(r.Body).Decode((&productVariant))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.svc.CreateProductVariant(r.Context(), productVariant)
	if err != nil {
		log.Println(err)
		switch {
		case errors.Is(err, ErrNameRequired),
			errors.Is(err, ErrInvalidPrice),
			errors.Is(err, ErrInvalidWholesale):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, ErrProductNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
