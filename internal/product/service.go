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
)

type productStore interface {
	ListProducts(ctx context.Context) ([]database.Product, error)
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

func (s *Service) ListProducts(ctx context.Context) ([]database.Product, error) {
	return s.repo.ListProducts(ctx)
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
