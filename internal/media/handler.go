package media

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ImageService interface {
	UploadImage(ctx context.Context, body io.Reader, params database.CreateProductImageParams) (database.ProductImage, error)
	DeleteImage(ctx context.Context, id int64) error
	GetListProductImages(ctx context.Context, productId int64) ([]database.ProductImage, error)
	GetListVariantImage(ctx context.Context, params database.ListVariantImagesParams) ([]database.ProductImage, error)
}

type Handler struct {
	svc ImageService
}

func NewHandler(svc ImageService) *Handler {
	return &Handler{svc: svc}
}

const MAX_UPLOAD_SIZE = 5 << 20 // 5MB

var errInvalidID = errors.New("id must be a positive integer")

func parseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidID
	}
	return id, nil
}

// /upload
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)      //  add wrapper with set limit 5 mb
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil { // wrapper throw the error if it exceed 5 mb while reading
		http.Error(w, "The uploaded file is too big. Please choose an file that's less than 5 mb in size", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buffer := make([]byte, 512) // to read first bytes and detect file type
	_, err = file.Read(buffer)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}

	contentType := http.DetectContentType(buffer)
	if _, ok := allowedFileTypes[contentType]; !ok {
		log.Println(err)
		http.Error(w, "File type is not allowed", http.StatusBadRequest)
		return
	}

	// check uploaded file includes the bytes read for type detection
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Println(err)
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	productId, err := parseID(r.FormValue("product_id"))
	if err != nil {
		http.Error(w, "Invalid product id", http.StatusBadRequest)
		return
	}

	// variant_id is optional: NULL means the image belongs to the whole product
	var variantId pgtype.Int8
	if raw := r.FormValue("variant_id"); raw != "" {
		id, err := parseID(raw)
		if err != nil {
			http.Error(w, "Invalid variant id", http.StatusBadRequest)
			return
		}
		variantId = pgtype.Int8{Int64: id, Valid: true}
	}

	productImageParams := database.CreateProductImageParams{
		ProductID:   productId,
		VariantID:   variantId,
		ContentType: contentType,
	}

	productImage, err := h.svc.UploadImage(r.Context(), file, productImageParams)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(productImage)
}

func (h *Handler) GetListProductImages(w http.ResponseWriter, r *http.Request) {
	productId, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid product id", http.StatusBadRequest)
		return
	}

	productImages, err := h.svc.GetListProductImages(r.Context(), productId)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(productImages)
}

func (h *Handler) GetListVariantImages(w http.ResponseWriter, r *http.Request) {
	productId, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid product id", http.StatusBadRequest)
		return
	}

	variantId, err := parseID(chi.URLParam(r, "variantId"))
	if err != nil {
		http.Error(w, "Invalid variant id", http.StatusBadRequest)
		return
	}

	variantImages, err := h.svc.GetListVariantImage(r.Context(), database.ListVariantImagesParams{
		ProductID: productId,
		VariantID: variantId,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(variantImages)
}

func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	imageId, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid image id", http.StatusBadRequest)
		return
	}

	if err := h.svc.DeleteImage(r.Context(), imageId); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrImageNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrForeignIdNotExists):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrKeyNotUnique):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
