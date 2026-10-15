package product

import (
	"context"
	"errors"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrSlugTaken        = errors.New("slug should be unique")
	ErrProductNotFound  = errors.New("product did not exist")
	ErrCategoryNotFound = errors.New("category did not exist")
)

type Repository struct {
	q *database.Queries
}

func NewRepository(q *database.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) ListProducts(ctx context.Context) ([]database.Product, error) {
	rows, err := r.q.ListProducts(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return []database.Product{}, pgErr
		}
		return []database.Product{}, err
	}

	return rows, nil
}

func (r *Repository) GetProductByID(ctx context.Context, id int64) (database.Product, error) {
	product, err := r.q.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.Product{}, ErrProductNotFound
		}
		return database.Product{}, err
	}

	return product, nil
}

func (r *Repository) CreateProduct(ctx context.Context, productParams database.CreateProductParams) error {
	err := r.q.CreateProduct(ctx, productParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique violation
				return ErrSlugTaken
			case "23503": // foreign key violation
				return ErrCategoryNotFound
			}
		}
		return err
	}

	return nil
}

func (r *Repository) ListProductVariants(ctx context.Context, productId int64) ([]database.ProductVariant, error) {
	productVariants, err := r.q.ListProductVariants(ctx, productId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return []database.ProductVariant{}, pgErr
		}
		return []database.ProductVariant{}, err
	}

	// encode an empty list as [] instead of null
	if productVariants == nil {
		return []database.ProductVariant{}, nil
	}

	return productVariants, nil
}

func (r *Repository) CreateProductVariant(ctx context.Context, variantParams database.CreateProductVariantParams) error {
	err := r.q.CreateProductVariant(ctx, variantParams)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23503 error code for foreign key violation
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrProductNotFound
		}
		return err
	}

	return nil
}
