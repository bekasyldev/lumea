package orders

import (
	"context"
	"errors"
	"testing"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

// fakeStore serves a fixed catalog and records what CreateOrder received
type fakeStore struct {
	orderStore
	variants    []database.GetVariantsForOrderRow
	gotOrder    database.CreateOrderParams
	gotItems    []database.CreateOrderItemParams
	statusCalls []database.UpdateOrderStatusParams
}

func variant(price int64, wholesale int64, minQty int32) database.GetVariantsForOrderRow {
	v := database.GetVariantsForOrderRow{Price: price}
	if minQty > 0 {
		v.WholesalePrice = pgtype.Int8{Int64: wholesale, Valid: true}
		v.WholesaleMinQty = pgtype.Int4{Int32: minQty, Valid: true}
	}
	return v
}

func (f *fakeStore) GetVariants(_ context.Context, ids []int64) ([]database.GetVariantsForOrderRow, error) {
	var out []database.GetVariantsForOrderRow
	for _, v := range f.variants {
		for _, id := range ids {
			if v.ID == id {
				out = append(out, v)
			}
		}
	}
	return out, nil
}

func (f *fakeStore) UpdateOrderStatus(_ context.Context, p database.UpdateOrderStatusParams) error {
	f.statusCalls = append(f.statusCalls, p)
	return nil
}

func (f *fakeStore) CreateOrder(_ context.Context, o database.CreateOrderParams, items []database.CreateOrderItemParams) (OrderDetails, error) {
	f.gotOrder, f.gotItems = o, items
	return OrderDetails{}, nil
}

var catalog = []database.GetVariantsForOrderRow{
	{ID: 1, ProductName: "Cream", VariantName: "50 ml", Price: 1000,
		WholesalePrice: pgtype.Int8{Int64: 800, Valid: true}, WholesaleMinQty: pgtype.Int4{Int32: 5, Valid: true}},
	{ID: 2, ProductName: "Serum", VariantName: "30 ml", Price: 3000},
}

func TestCreateOrderPricesOnServer(t *testing.T) {
	store := &fakeStore{variants: catalog}
	svc := NewService(store)

	_, err := svc.CreateOrder(context.Background(), CreateOrderInput{
		CustomerName:  "Aida",
		CustomerPhone: "+77000000000",
		Items: []OrderItemInput{
			{ProductVariantID: 1, Quantity: 5}, // reaches wholesale minimum
			{ProductVariantID: 2, Quantity: 2}, // no wholesale price
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := store.gotItems[0]; got.UnitPrice != 800 || got.ProductName != "Cream" || got.VariantName != "50 ml" {
		t.Errorf("item 1 = %+v, want wholesale price 800 and catalog names", got)
	}
	if got := store.gotItems[1].UnitPrice; got != 3000 {
		t.Errorf("item 2 unit price = %d, want 3000", got)
	}
	if got := store.gotOrder.TotalPrice; got != 5*800+2*3000 {
		t.Errorf("total = %d, want %d", got, 5*800+2*3000)
	}
}

func TestCreateOrderBelowWholesaleMinimum(t *testing.T) {
	store := &fakeStore{variants: catalog}
	_, err := NewService(store).CreateOrder(context.Background(), CreateOrderInput{
		CustomerName: "Aida", CustomerPhone: "1",
		Items: []OrderItemInput{{ProductVariantID: 1, Quantity: 4}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := store.gotItems[0].UnitPrice; got != 1000 {
		t.Errorf("unit price = %d, want regular price 1000", got)
	}
}

func TestCreateOrderRejectsBadItems(t *testing.T) {
	tests := []struct {
		name  string
		items []OrderItemInput
		want  error
	}{
		{"unknown or inactive variant", []OrderItemInput{{ProductVariantID: 99, Quantity: 1}}, ErrVariantNotFound},
		{"duplicate variant", []OrderItemInput{{ProductVariantID: 1, Quantity: 1}, {ProductVariantID: 1, Quantity: 2}}, ErrDuplicateVariant},
		{"zero quantity", []OrderItemInput{{ProductVariantID: 1, Quantity: 0}}, ErrInvalidQuantity},
		{"no items", nil, ErrItemsRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(&fakeStore{variants: catalog}).CreateOrder(context.Background(), CreateOrderInput{
				CustomerName: "Aida", CustomerPhone: "1", Items: tt.items,
			})
			if !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestUnitPrice(t *testing.T) {
	tests := []struct {
		name string
		v    database.GetVariantsForOrderRow
		qty  int32
		want int64
	}{
		{"no wholesale pricing", variant(600, 0, 0), 1000, 600},
		{"min qty not valid, even if qty is high", database.GetVariantsForOrderRow{Price: 600, WholesaleMinQty: pgtype.Int4{Int32: 30, Valid: false}}, 20, 600},
		{"below minimum", variant(600, 200, 8), 7, 600},
		{"exactly at minimum", variant(600, 200, 8), 8, 200},
		{"above minimum", variant(600, 200, 8), 50, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unitPrice(tt.v, tt.qty); got != tt.want {
				t.Errorf("unitPrice() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestUpdateOrderStatus(t *testing.T) {
	tests := []struct {
		status  string
		wantErr error
	}{
		{"pending", nil},
		{"confirmed", nil},
		{"shipped", nil},
		{"delivered", nil},
		{"cancelled", nil},
		{"refunded", ErrInvalidStatus},
		{"", ErrInvalidStatus},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			store := &fakeStore{}
			err := NewService(store).UpdateOrderStatus(context.Background(),
				database.UpdateOrderStatusParams{ID: 1, Status: tt.status})

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			wantCalls := 0
			if tt.wantErr == nil {
				wantCalls = 1
			}
			if len(store.statusCalls) != wantCalls {
				t.Errorf("store calls = %d, want %d", len(store.statusCalls), wantCalls)
			}
		})
	}
}
