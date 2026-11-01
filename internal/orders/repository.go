package orders

import (
	"context"
	"errors"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrOrderNotFound     = errors.New("order did not exist")
	ErrOrderItemNotFound = errors.New("order item did not exist")
	ErrVariantNotFound   = errors.New("product variant did not exist")
	ErrDuplicateVariant  = errors.New("product variant appears more than once in order")
	ErrLastOrderItem     = errors.New("cannot delete the last item of an order, delete the order instead")
	ErrInsufficientStock = errors.New("stock number is less than order quantity")
)

// for transaction operations(all or nothing)
type txBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	db txBeginner
	q  *database.Queries
}

func NewRepository(db txBeginner, q *database.Queries) *Repository {
	return &Repository{db: db, q: q}
}

func (r *Repository) ListOrders(ctx context.Context) ([]database.Order, error) {
	orders, err := r.q.ListOrders(ctx)
	if err != nil {
		return nil, err
	}

	// encode an empty list as [] instead of null
	if orders == nil {
		return []database.Order{}, nil
	}

	return orders, nil
}

func (r *Repository) GetOrderByID(ctx context.Context, id int64) (database.Order, error) {
	order, err := r.q.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.Order{}, ErrOrderNotFound
		}
		return database.Order{}, err
	}

	return order, nil
}

func (r *Repository) ListOrderItems(ctx context.Context, orderId int64) ([]database.OrderItem, error) {
	items, err := r.q.ListOrderItemsByOrderID(ctx, orderId)
	if err != nil {
		return nil, err
	}

	if items == nil {
		return []database.OrderItem{}, nil
	}

	return items, nil
}

func (r *Repository) GetVariants(ctx context.Context, ids []int64) ([]database.GetVariantsForOrderRow, error) {
	return r.q.GetVariantsForOrder(ctx, ids)
}

func (r *Repository) CreateOrder(ctx context.Context, orderParams database.CreateOrderParams, items []database.CreateOrderItemParams) (OrderDetails, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return OrderDetails{}, err
	}
	// to undo changes
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	// atomic update
	for _, orderItem := range items {
		var params = database.DecreaseVariantStockParams{ID: orderItem.ProductVariantID, Stock: orderItem.Quantity}
		rows, err := qtx.DecreaseVariantStock(ctx, params)
		if err != nil {
			return OrderDetails{}, err
		}

		if rows == 0 {
			return OrderDetails{}, ErrInsufficientStock
		}
	}

	order, err := qtx.CreateOrder(ctx, orderParams)
	if err != nil {
		return OrderDetails{}, err
	}

	created := make([]database.OrderItem, 0, len(items))
	for _, item := range items {
		item.OrderID = order.ID
		row, err := qtx.CreateOrderItem(ctx, item)
		if err != nil {
			return OrderDetails{}, mapItemErr(err)
		}
		created = append(created, row)
	}

	if err := tx.Commit(ctx); err != nil {
		return OrderDetails{}, err
	}

	return OrderDetails{Order: order, Items: created}, nil
}

func (r *Repository) UpdateOrder(ctx context.Context, orderParams database.UpdateOrderParams) error {
	rows, err := r.q.UpdateOrder(ctx, orderParams)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOrderNotFound
	}

	return nil
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, statusParams database.UpdateOrderStatusParams) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	// lock the order row
	if _, err := qtx.LockOrderForUpdate(ctx, statusParams.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderNotFound
		}
		return err
	}

	order, err := qtx.GetOrderByID(ctx, statusParams.ID)
	if err != nil {
		return err
	}

	// oldStatus == newStatus
	if OrderStatus(statusParams.Status) == OrderStatus(order.Status) {
		return nil
	}

	// if status moved from cancelled to another then decrease stock
	if OrderStatus(order.Status) == StatusCancelled {
		orderItems, err := qtx.ListOrderItemsByOrderID(ctx, statusParams.ID)
		if err != nil {
			return err
		}

		for _, o := range orderItems {
			var params = database.DecreaseVariantStockParams{ID: o.ProductVariantID, Stock: o.Quantity}
			rows, err := qtx.DecreaseVariantStock(ctx, params)
			if err != nil {
				return err
			}

			if rows == 0 {
				return ErrInsufficientStock
			}
		}
	}

	rows, err := qtx.UpdateOrderStatus(ctx, statusParams)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOrderNotFound
	}

	// return to the stock if status cancelled
	if OrderStatus(statusParams.Status) == StatusCancelled {
		ordersItems, err := qtx.ListOrderItemsByOrderID(ctx, statusParams.ID)
		if err != nil {
			return err
		}

		for _, o := range ordersItems {
			var params = database.ReturnVariantToStockParams{ID: o.ProductVariantID, Stock: o.Quantity}
			if err := qtx.ReturnVariantToStock(ctx, params); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteOrderItem(ctx context.Context, itemParams database.DeleteOrderItemParams) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)

	// lock the order row
	if _, err := qtx.LockOrderForUpdate(ctx, itemParams.OrderID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderNotFound
		}
		return err
	}

	orderItem, err := qtx.GetOrderItemByID(ctx, itemParams.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderItemNotFound
		}
		return err
	}

	if orderItem.OrderID != itemParams.OrderID {
		return ErrOrderItemNotFound
	}

	order, err := qtx.GetOrderByID(ctx, itemParams.OrderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderNotFound
		}

		return err
	}

	// if order cancelled it already retunred variant to the stock
	if OrderStatus(order.Status) != StatusCancelled {
		var params = database.ReturnVariantToStockParams{ID: orderItem.ProductVariantID, Stock: orderItem.Quantity}
		if err := qtx.ReturnVariantToStock(ctx, params); err != nil {
			return err
		}
	}

	rows, err := qtx.DeleteOrderItem(ctx, itemParams)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOrderItemNotFound
	}

	// order must keep at least one item
	remaining, err := qtx.CountOrderItems(ctx, itemParams.OrderID)
	if err != nil {
		return err
	}
	if remaining == 0 {
		return ErrLastOrderItem
	}

	if _, err := qtx.RecalculateOrderTotal(ctx, itemParams.OrderID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) DeleteOrder(ctx context.Context, id int64) error {
	rows, err := r.q.DeleteOrder(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrOrderNotFound
	}

	return nil
}

func mapItemErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		// 23503 error code for foreign key violation
		case "order_items_order_id_fkey":
			return ErrOrderNotFound
		case "order_items_product_variant_id_fkey":
			return ErrVariantNotFound
		// 23505 error code for unique violation
		case "order_items_order_variant_unique":
			return ErrDuplicateVariant
		}
	}
	return err
}
