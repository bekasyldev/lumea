package media

import (
	"context"
	"errors"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrKeyNotUnique       = errors.New("s3 key should be unique")
	ErrForeignIdNotExists = errors.New("product or variant did not exist")
	ErrImageNotFound      = errors.New("image did not exist")
)

type Repository struct {
	q *database.Queries
}

func NewRepository(q *database.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) Create(ctx context.Context, params database.CreateProductImageParams) (database.ProductImage, error) {
	productImage, err := r.q.CreateProductImage(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique violation
				return database.ProductImage{}, ErrKeyNotUnique
			case "23503": // foreign key violation
				return database.ProductImage{}, ErrForeignIdNotExists
			}
		}
		return database.ProductImage{}, err
	}

	return productImage, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (database.ProductImage, error) {
	productImage, err := r.q.GetProductImageByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.ProductImage{}, ErrImageNotFound
		}
		return database.ProductImage{}, err
	}

	return productImage, nil
}

func (r *Repository) ListByProduct(ctx context.Context, productId int64) ([]database.ProductImage, error) {
	productImages, err := r.q.ListProductImages(ctx, productId)
	if err != nil {
		return []database.ProductImage{}, err
	}

	// encode an empty list as [] instead of null
	if productImages == nil {
		return []database.ProductImage{}, nil
	}

	return productImages, nil
}

func (r *Repository) ListByVariant(ctx context.Context, params database.ListVariantImagesParams) ([]database.ProductImage, error) {
	variantImages, err := r.q.ListVariantImages(ctx, params)
	if err != nil {
		return []database.ProductImage{}, err
	}

	// encode an empty list as [] instead of null
	if variantImages == nil {
		return []database.ProductImage{}, nil
	}

	return variantImages, nil
}

func (r *Repository) GetMain(ctx context.Context, productId int64) (database.ProductImage, error) {
	productImage, err := r.q.GetMainProductImage(ctx, productId)
	if err != nil {
		// product has no images yet
		if errors.Is(err, pgx.ErrNoRows) {
			return database.ProductImage{}, ErrImageNotFound
		}
		return database.ProductImage{}, err
	}

	return productImage, nil
}

func (r *Repository) UpdatePosition(ctx context.Context, params database.UpdateProductImagePositionParams) error {
	rows, err := r.q.UpdateProductImagePosition(ctx, params)
	if err != nil {
		return err
	}

	// UPDATE doesn't fail on a missing id, it just touches 0 rows
	if rows == 0 {
		return ErrImageNotFound
	}

	return nil
}

// Delete removes the row and returns its s3 key
func (r *Repository) Delete(ctx context.Context, id int64) (string, error) {
	key, err := r.q.DeleteProductImage(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrImageNotFound
		}
		return "", err
	}

	return key, nil
}

func (r *Repository) ListKeysByProduct(ctx context.Context, productId int64) ([]string, error) {
	keys, err := r.q.ListProductImageKeys(ctx, productId)
	if err != nil {
		return []string{}, err
	}

	return keys, nil
}
