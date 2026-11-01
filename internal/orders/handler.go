package orders

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

type OrderService interface {
	ListOrders(ctx context.Context) ([]database.Order, error)
	GetOrderByID(ctx context.Context, id int64) (OrderDetails, error)
	CreateOrder(ctx context.Context, input CreateOrderInput) (OrderDetails, error)
	UpdateOrder(ctx context.Context, orderParams database.UpdateOrderParams) error
	UpdateOrderStatus(ctx context.Context, statusParams database.UpdateOrderStatusParams) error
	DeleteOrderItem(ctx context.Context, itemParams database.DeleteOrderItemParams) error
	DeleteOrder(ctx context.Context, id int64) error
}

type Handler struct {
	svc OrderService
}

func NewHandler(svc OrderService) *Handler {
	return &Handler{svc: svc}
}

// GET /orders
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.svc.ListOrders(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

// GET /orders/{id}
func (h *Handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderId, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	order, err := h.svc.GetOrderByID(r.Context(), orderId)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

// POST /orders
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var input CreateOrderInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := h.svc.CreateOrder(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

// PUT /orders/{id}
func (h *Handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	orderId, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	var orderParams database.UpdateOrderParams

	err := json.NewDecoder(r.Body).Decode(&orderParams)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// id always comes from the url, never from the body
	orderParams.ID = orderId

	err = h.svc.UpdateOrder(r.Context(), orderParams)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /orders/{id}/status
func (h *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderId, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	var statusParams database.UpdateOrderStatusParams

	err := json.NewDecoder(r.Body).Decode(&statusParams)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	statusParams.ID = orderId

	err = h.svc.UpdateOrderStatus(r.Context(), statusParams)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /orders/{id}
func (h *Handler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	orderId, ok := parseID(w, r, "id")
	if !ok {
		return
	}

	err := h.svc.DeleteOrder(r.Context(), orderId)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /orders/{id}/items/{itemId}
func (h *Handler) DeleteOrderItem(w http.ResponseWriter, r *http.Request) {
	orderId, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	itemId, ok := parseID(w, r, "itemId")
	if !ok {
		return
	}

	err := h.svc.DeleteOrderItem(r.Context(), database.DeleteOrderItemParams{ID: itemId, OrderID: orderId})
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseID reads an int64 url param, writes 400 and returns false if it is invalid
func parseID(w http.ResponseWriter, r *http.Request, param string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, param), 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid "+param, http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

// writeError maps domain errors to http status codes
func writeError(w http.ResponseWriter, err error) {
	log.Println(err)
	switch {
	case errors.Is(err, ErrCustomerNameRequired),
		errors.Is(err, ErrCustomerPhoneRequired),
		errors.Is(err, ErrItemsRequired),
		errors.Is(err, ErrInvalidQuantity),
		errors.Is(err, ErrInvalidStatus),
		errors.Is(err, ErrDuplicateVariant):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrOrderNotFound),
		errors.Is(err, ErrOrderItemNotFound),
		errors.Is(err, ErrVariantNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, ErrLastOrderItem):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, ErrInsufficientStock):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
