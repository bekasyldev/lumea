//go:build integration

package orders

import (
	"context"
	"errors"
	"sync"
	"testing"

	database "github.com/bekasyldev/lumea/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("lumea-test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	// runs in LIFO order after the test, pass or fail.
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable", "pool_max_conns=25")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(pool.Close)

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.Up(db, "../../db/migrations"); err != nil {
		t.Fatalf("migration: %v", err)
	}

	return pool
}

type Product struct {
	name        string
	slug        string
	description string
	stock       int32
	categoryID  int64
}

type Category struct {
	name string
	slug string
}

func seedCategory(t *testing.T, pool *pgxpool.Pool, c Category) int64 {
	t.Helper()
	ctx := context.Background()

	var categoryID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO categories (name, slug) VALUES ($1, $2) RETURNING id`,
		c.name, c.slug,
	).Scan(&categoryID)
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}

	return categoryID
}

// seedVariant inserts product and variant with the given stock
func seedVariant(t *testing.T, pool *pgxpool.Pool, p Product) int64 {
	t.Helper()
	ctx := context.Background()

	var productID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO products (name, slug, description, category_id)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		p.name, p.slug, p.description, p.categoryID,
	).Scan(&productID)
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}

	var variantID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO product_variants
		   (product_id, name, price, wholesale_price, wholesale_min_qty, stock)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		productID, "50 ml", 1000, 800, 5, p.stock,
	).Scan(&variantID)
	if err != nil {
		t.Fatalf("seed variant: %v", err)
	}

	return variantID
}

// variantStock reads the current stock
func variantStock(t *testing.T, pool *pgxpool.Pool, variantID int64) int32 {
	t.Helper()
	var stock int32
	err := pool.QueryRow(context.Background(),
		`SELECT stock FROM product_variants WHERE id = $1`, variantID,
	).Scan(&stock)
	if err != nil {
		t.Fatalf("read stock: %v", err)
	}
	return stock
}

func TestRaceCondtionOrders(t *testing.T) {
	pool := setupPostgres(t)
	repo := NewRepository(pool, database.New(pool))
	svc := NewService(repo)

	c := Category{
		name: "Skincare",
		slug: "skincare",
	}

	categoryID := seedCategory(t, pool, c)

	p := Product{
		name:        "Cream",
		slug:        "cream",
		description: "Moisturizing face cream",
		stock:       10,
		categoryID:  categoryID,
	}

	variantID := seedVariant(t, pool, p)
	input := CreateOrderInput{
		CustomerName:  "Aida",
		CustomerPhone: "+77000000000",
		Items:         []OrderItemInput{{ProductVariantID: variantID, Quantity: 1}},
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.CreateOrder(context.Background(), input)
			if err != nil {
				errCh <- err
			}
		}()

	}

	wg.Wait()
	close(errCh)

	fails := 0
	for err := range errCh {
		if !errors.Is(err, ErrInsufficientStock) {
			t.Fatalf("unexpected error: %v", err)
		}
		fails += 1
	}

	if successes := 20 - fails; successes != 10 {
		t.Errorf("successful orders = %d, want 10", successes)
	}
	if got := variantStock(t, pool, variantID); got != 0 {
		t.Errorf("stock = %d, want 0", got)
	}

	var orderCount int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM orders`).Scan(&orderCount); err != nil {
		t.Fatal(err)
	}
	if orderCount != 10 {
		t.Errorf("orders in db = %d, want 10", orderCount)
	}

}

func TestTransitionRollbaackOrder(t *testing.T) {
	pool := setupPostgres(t)
	repo := NewRepository(pool, database.New(pool))
	svc := NewService(repo)

	c := Category{
		name: "Skincare",
		slug: "skincare",
	}
	categoryID := seedCategory(t, pool, c)

	p1 := Product{
		name:        "Cream",
		slug:        "cream",
		description: "Moisturizing face cream",
		stock:       10,
		categoryID:  categoryID,
	}

	p2 := Product{
		name:        "Soap",
		slug:        "soap",
		description: "Universal soap for all skinks)",
		stock:       1,
		categoryID:  categoryID,
	}

	v1ID := seedVariant(t, pool, p1)
	v2ID := seedVariant(t, pool, p2)

	input := CreateOrderInput{
		CustomerName:  "Aida",
		CustomerPhone: "+77000000000",
		Items: []OrderItemInput{
			{ProductVariantID: v1ID, Quantity: 3},
			{ProductVariantID: v2ID, Quantity: 5}, // not have enough stock
		},
	}

	_, err := svc.CreateOrder(context.Background(), input)
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatal(err)
	}

	if got := variantStock(t, pool, v1ID); got != 10 {
		t.Fatalf("err = %v ,want %v", err, ErrInsufficientStock) // should rollback
	}

	if got := variantStock(t, pool, v2ID); got != 1 {
		t.Fatalf("err = %v ,want %v", err, ErrInsufficientStock) // should rollback
	}
}
