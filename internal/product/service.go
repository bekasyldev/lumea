package product

import (
	"context"
	"errors"
	"strings"

	database "github.com/bekasyldev/lumea/internal/database"
)

var (
	ErrNameRequired     = errors.New("name is required")
	ErrSlugRequired     = errors.New("slug is required")
	ErrInvalidPrice     = errors.New("price must be greater than 0")
	ErrInvalidWholesale = errors.New("wholesale price and min qty must be set together, price must be lower than regular price and min qty at least 2")
	ErrInvalidLimit     = errors.New("limit should be between 1 and 100")
	ErrInvalidDir       = errors.New("dir should be equal to 'next' or 'prev'")
)

type productStore interface {
	ListProducts(ctx context.Context, cursor int64, dir string, limit int32) ([]database.Product, error)
	GetProductByID(ctx context.Context, id int64) (database.Product, error)
	CreateProduct(ctx context.Context, productParams database.CreateProductParams) error
	ListProductVariants(ctx context.Context, productId int64) ([]database.ProductVariant, error)
	CreateProductVariant(ctx context.Context, variantParams database.CreateProductVariantParams) error
}

type Service struct {
	repo productStore
}

func NewService(repo productStore) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListProducts(ctx context.Context, curs int64, dir string, limit int32) ([]database.Product, int64, bool, error) {
	if limit < 1 || limit > 100 {
		return []database.Product{}, 0, false, ErrInvalidLimit
	}

	if dir != string(DirectionNext) && dir != string(DirectionPrev) {
		return []database.Product{}, 0, false, ErrInvalidDir
	}

	rows, err := s.repo.ListProducts(ctx, curs, dir, limit+1)
	if err != nil {
		return []database.Product{}, 0, false, err
	}

	hasMore := int32(len(rows)) > limit
	if hasMore {
		if dir == string(DirectionNext) {
			rows = rows[:limit]
		} else {
			rows = rows[1:]
		}
	}
	if len(rows) == 0 {
		return []database.Product{}, curs, false, nil
	}

	cursor := rows[len(rows)-1].ID
	if dir == string(DirectionPrev) {
		cursor = rows[0].ID
	}

	return rows, cursor, hasMore, nil
}

func (s *Service) CreateProduct(ctx context.Context, productParams database.CreateProductParams) error {
	productParams.Name = strings.TrimSpace(productParams.Name)
	productParams.Slug = strings.ToLower(strings.TrimSpace(productParams.Slug))
	productParams.Description = strings.TrimSpace(productParams.Description)

	if productParams.Name == "" {
		return ErrNameRequired
	}
	if productParams.Slug == "" {
		return ErrSlugRequired
	}

	return s.repo.CreateProduct(ctx, productParams)
}

func (s *Service) ListProductVariants(ctx context.Context, productId int64) ([]database.ProductVariant, error) {
	// check if product exist
	if _, err := s.repo.GetProductByID(ctx, productId); err != nil {
		return nil, err
	}

	return s.repo.ListProductVariants(ctx, productId)
}

func (s *Service) CreateProductVariant(ctx context.Context, variantParams database.CreateProductVariantParams) error {
	variantParams.Name = strings.TrimSpace(variantParams.Name)

	if variantParams.Name == "" {
		return ErrNameRequired
	}
	if variantParams.Price <= 0 {
		return ErrInvalidPrice
	}

	// mirror the wholesale_pair / wholesale_cheaper DB constraints
	wp, wq := variantParams.WholesalePrice, variantParams.WholesaleMinQty
	if wp.Valid != wq.Valid {
		return ErrInvalidWholesale
	}
	if wp.Valid && (wp.Int64 <= 0 || wp.Int64 >= variantParams.Price || wq.Int32 < 2) {
		return ErrInvalidWholesale
	}

	return s.repo.CreateProductVariant(ctx, variantParams)
}
