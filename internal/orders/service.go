package orders

import (
	"context"
	"errors"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrCustomerNameRequired  = errors.New("customer name is required")
	ErrCustomerPhoneRequired = errors.New("customer phone is required")
	ErrItemsRequired         = errors.New("order must have at least one item")
	ErrInvalidQuantity       = errors.New("quantity must be greater than 0")
	ErrInvalidStatus         = errors.New("invalid order status")
)

type CreateOrderInput struct {
	CustomerName  string
	CustomerPhone string
	Comment       string
	Items         []OrderItemInput
}

// names and prices are snapshotted from the catalog on the server, never taken from the client
type OrderItemInput struct {
	ProductVariantID int64
	Quantity         int32
}

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

type OrderDetails struct {
	database.Order
	Items []database.OrderItem
}

type orderStore interface {
	ListOrders(ctx context.Context) ([]database.Order, error)
	GetOrderByID(ctx context.Context, id int64) (database.Order, error)
	ListOrderItems(ctx context.Context, orderId int64) ([]database.OrderItem, error)
	GetVariants(ctx context.Context, ids []int64) ([]database.GetVariantsForOrderRow, error)
	CreateOrder(ctx context.Context, orderParams database.CreateOrderParams, items []database.CreateOrderItemParams) (OrderDetails, error)
	UpdateOrder(ctx context.Context, orderParams database.UpdateOrderParams) error
	UpdateOrderStatus(ctx context.Context, statusParams database.UpdateOrderStatusParams) error
	DeleteOrderItem(ctx context.Context, itemParams database.DeleteOrderItemParams) error
	DeleteOrder(ctx context.Context, id int64) error
}

type Service struct {
	repo orderStore
}

func NewService(repo orderStore) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListOrders(ctx context.Context) ([]database.Order, error) {
	return s.repo.ListOrders(ctx)
}

func (s *Service) GetOrderByID(ctx context.Context, id int64) (OrderDetails, error) {
	order, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return OrderDetails{}, err
	}

	items, err := s.repo.ListOrderItems(ctx, id)
	if err != nil {
		return OrderDetails{}, err
	}

	return OrderDetails{Order: order, Items: items}, nil
}

func (s *Service) CreateOrder(ctx context.Context, input CreateOrderInput) (OrderDetails, error) {
	// validation
	if input.CustomerName == "" {
		return OrderDetails{}, ErrCustomerNameRequired
	}

	if input.CustomerPhone == "" {
		return OrderDetails{}, ErrCustomerPhoneRequired
	}

	if len(input.Items) == 0 {
		return OrderDetails{}, ErrItemsRequired
	}

	ids := make([]int64, 0, len(input.Items))
	seen := make(map[int64]bool, len(input.Items))
	for _, item := range input.Items {
		if item.Quantity < 1 {
			return OrderDetails{}, ErrInvalidQuantity
		}
		if seen[item.ProductVariantID] {
			return OrderDetails{}, ErrDuplicateVariant
		}
		seen[item.ProductVariantID] = true
		ids = append(ids, item.ProductVariantID)
	}

	variants, err := s.repo.GetVariants(ctx, ids)
	if err != nil {
		return OrderDetails{}, err
	}

	// missing or inactive variants are not returned
	byID := make(map[int64]database.GetVariantsForOrderRow, len(variants))
	for _, v := range variants {
		byID[v.ID] = v
	}

	var total int64
	items := make([]database.CreateOrderItemParams, 0, len(input.Items))
	for _, item := range input.Items {
		v, ok := byID[item.ProductVariantID]
		if !ok {
			return OrderDetails{}, ErrVariantNotFound
		}

		price := unitPrice(v, item.Quantity)
		total += price * int64(item.Quantity)

		items = append(items, database.CreateOrderItemParams{
			ProductVariantID: v.ID,
			ProductName:      v.ProductName,
			VariantName:      v.VariantName,
			Quantity:         item.Quantity,
			UnitPrice:        price,
		})
	}

	orderParams := database.CreateOrderParams{
		CustomerName:  input.CustomerName,
		CustomerPhone: input.CustomerPhone,
		Comment:       pgtype.Text{String: input.Comment, Valid: input.Comment != ""},
		TotalPrice:    total,
	}

	return s.repo.CreateOrder(ctx, orderParams, items)
}

// unitPrice applies the wholesale rule: wholesale price once quantity reaches the minimum
func unitPrice(v database.GetVariantsForOrderRow, qty int32) int64 {
	if v.WholesaleMinQty.Valid && qty >= v.WholesaleMinQty.Int32 {
		return v.WholesalePrice.Int64
	}
	return v.Price
}

func (s *Service) UpdateOrder(ctx context.Context, orderParams database.UpdateOrderParams) error {
	return s.repo.UpdateOrder(ctx, orderParams)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, statusParams database.UpdateOrderStatusParams) error {
	switch OrderStatus(statusParams.Status) {
	case StatusPending, StatusConfirmed, StatusShipped, StatusDelivered, StatusCancelled:
		// Valid status
	default:
		return ErrInvalidStatus
	}

	// repo returns ErrOrderNotFound when no row matched
	return s.repo.UpdateOrderStatus(ctx, statusParams)
}

func (s *Service) DeleteOrderItem(ctx context.Context, itemParams database.DeleteOrderItemParams) error {
	return s.repo.DeleteOrderItem(ctx, itemParams)
}

func (s *Service) DeleteOrder(ctx context.Context, id int64) error {
	return s.repo.DeleteOrder(ctx, id)
}
