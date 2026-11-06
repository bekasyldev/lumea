package product

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/bekasyldev/lumea/internal/json"
	"github.com/go-chi/chi/v5"
)

type ProductService interface {
	ListProducts(ctx context.Context, cursor int64, dir string, limit int32) ([]database.Product, int64, bool, error)
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
	ctx := r.Context()
	q := r.URL.Query()

	// defaults: first page, going forward, 20 items
	dir := q.Get("dir")
	if dir == "" {
		dir = string(DirectionNext)
	}

	limit := int32(20)
	if s := q.Get("limit"); s != "" {
		l, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = int32(l)
	}

	// no cursor: next starts from the beginning, prev starts from the end
	var cursor int64 // equal to zero for next direction
	if dir == string(DirectionPrev) {
		cursor = math.MaxInt64
	}
	if s := q.Get("cursor"); s != "" {
		c, err := strconv.ParseInt(s, 10, 64)
		if err != nil || c < 0 {
			http.Error(w, "invalid cursor", http.StatusBadRequest)
			return
		}
		cursor = c
	}

	products, nextCursor, hasMore, err := h.svc.ListProducts(ctx, cursor, dir, limit)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidLimit), errors.Is(err, ErrInvalidDir):
			slog.WarnContext(ctx, "product: invalid list params", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			slog.ErrorContext(ctx, "product: list failed", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusOK, Pagination{
		Data:    products,
		Cursor:  nextCursor,
		HasMore: hasMore,
	})
}

// POST /products
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var product database.CreateProductParams

	parseErr := json.Read(r, &product)
	if parseErr != nil {
		slog.WarnContext(ctx, "product: invalid request body", "err", parseErr)
		http.Error(w, parseErr.Error(), http.StatusBadRequest)
		return
	}

	err := h.svc.CreateProduct(ctx, product)
	if err != nil {
		switch {
		case errors.Is(err, ErrNameRequired) || errors.Is(err, ErrSlugRequired):
			slog.WarnContext(ctx, "product: name or slug missing", "err", err)
			http.Error(w, "invalid credentials", http.StatusBadRequest)
			return
		case errors.Is(err, ErrSlugTaken):
			slog.WarnContext(ctx, "product: slug taken", "err", err, "slug", product.Slug)
			http.Error(w, "slug alreday in use", http.StatusBadRequest)
			return
		case errors.Is(err, ErrCategoryNotFound):
			slog.WarnContext(ctx, "product: category not found", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		default:
			slog.ErrorContext(ctx, "product: create failed", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

}

// GET /products/{id}/variants
func (h *Handler) ListProductVariants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	productId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		slog.WarnContext(ctx, "product: invalid product id", "err", err)
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	productVariants, err := h.svc.ListProductVariants(ctx, productId)
	if err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			slog.WarnContext(ctx, "product: not found", "err", err, "product_id", productId)
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			slog.ErrorContext(ctx, "product: list variants failed", "err", err, "product_id", productId)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	json.Write(w, http.StatusOK, productVariants)
}

// POST /products/{id}/variants
func (h *Handler) CreateProductVariant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	_, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		slog.WarnContext(ctx, "product: invalid product id", "err", err)
		http.Error(w, "invalid product id", http.StatusBadRequest)
		return
	}

	var productVariant database.CreateProductVariantParams

	err = json.Read(r, &productVariant)
	if err != nil {
		slog.WarnContext(ctx, "product: invalid request body", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.svc.CreateProductVariant(ctx, productVariant)
	if err != nil {
		switch {
		case errors.Is(err, ErrNameRequired),
			errors.Is(err, ErrInvalidPrice),
			errors.Is(err, ErrInvalidWholesale):
			slog.WarnContext(ctx, "product: invalid variant", "err", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, ErrProductNotFound):
			slog.WarnContext(ctx, "product: not found", "err", err)
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			slog.ErrorContext(ctx, "product: create variant failed", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
