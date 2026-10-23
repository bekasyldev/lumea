package media

import (
	"context"
	"fmt"
	"io"

	database "github.com/bekasyldev/lumea/internal/database"

	"github.com/google/uuid"
)

type fileStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Delete(ctx context.Context, key string) error
	URL(key string) string
}

type Service struct {
	storage fileStorage
	repo    mediaStore
}

func NewService(storage fileStorage, repo mediaStore) *Service {
	return &Service{storage: storage, repo: repo}
}

type mediaStore interface {
	Create(ctx context.Context, params database.CreateProductImageParams) (database.ProductImage, error)
	GetByID(ctx context.Context, id int64) (database.ProductImage, error)
	ListByProduct(ctx context.Context, productId int64) ([]database.ProductImage, error)
	ListByVariant(ctx context.Context, params database.ListVariantImagesParams) ([]database.ProductImage, error)
	GetMain(ctx context.Context, productId int64) (database.ProductImage, error)
	UpdatePosition(ctx context.Context, params database.UpdateProductImagePositionParams) error
	Delete(ctx context.Context, id int64) (string, error)
	ListKeysByProduct(ctx context.Context, productId int64) ([]string, error)
}

var allowedFileTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func extFor(contentType string) string {
	return allowedFileTypes[contentType]
}

func (s *Service) UploadImage(ctx context.Context, body io.Reader, params database.CreateProductImageParams) (database.ProductImage, error) {
	key := fmt.Sprintf("images/%s%s", uuid.NewString(), extFor(params.ContentType))
	if err := s.storage.Upload(ctx, key, body, params.ContentType); err != nil {
		return database.ProductImage{}, err
	}

	params.S3Key = key
	m, err := s.repo.Create(ctx, params)
	if err != nil {
		_ = s.storage.Delete(ctx, key) // DB failed so need to be removed from s3 too
		return database.ProductImage{}, err
	}

	return m, nil
}

func (s *Service) DeleteImage(ctx context.Context, id int64) error {
	key, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return s.storage.Delete(ctx, key)
}

func (s *Service) GetListProductImages(ctx context.Context, productId int64) ([]database.ProductImage, error) {
	productImages, err := s.repo.ListByProduct(ctx, productId)
	if err != nil {
		return []database.ProductImage{}, err
	}

	return productImages, nil
}

func (s *Service) GetListVariantImage(ctx context.Context, params database.ListVariantImagesParams) ([]database.ProductImage, error) {
	variantImages, err := s.repo.ListByVariant(ctx, params)
	if err != nil {
		return []database.ProductImage{}, err
	}

	return variantImages, err
}
