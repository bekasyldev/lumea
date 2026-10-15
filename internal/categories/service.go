package categories

import (
	"context"
	"errors"
	"strings"

	database "github.com/bekasyldev/lumea/internal/database"
)

var (
	ErrNameRequired = errors.New("name is required")
	ErrSlugRequired = errors.New("slug is required")
)

type categoryStore interface {
	ListCategories(ctx context.Context) ([]database.Category, error)
	GetCategoryByID(ctx context.Context, id int64) (database.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (database.Category, error)
	CreateCategory(ctx context.Context, categoryParams database.CreateCategoryParams) (database.Category, error)
	UpdateCategory(ctx context.Context, categoryParams database.UpdateCategoryParams) error
	DeleteCategory(ctx context.Context, id int64) error
}

type Service struct {
	repo categoryStore
}

func NewService(repo categoryStore) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListCategories(ctx context.Context) ([]database.Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *Service) GetCategoryByID(ctx context.Context, id int64) (database.Category, error) {
	return s.repo.GetCategoryByID(ctx, id)
}

func (s *Service) GetCategoryBySlug(ctx context.Context, slug string) (database.Category, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))

	return s.repo.GetCategoryBySlug(ctx, slug)
}

func (s *Service) CreateCategory(ctx context.Context, categoryParams database.CreateCategoryParams) (database.Category, error) {
	var err error
	categoryParams.Name, categoryParams.Slug, err = normalize(categoryParams.Name, categoryParams.Slug)
	if err != nil {
		return database.Category{}, err
	}

	return s.repo.CreateCategory(ctx, categoryParams)
}

func (s *Service) UpdateCategory(ctx context.Context, categoryParams database.UpdateCategoryParams) error {
	var err error
	categoryParams.Name, categoryParams.Slug, err = normalize(categoryParams.Name, categoryParams.Slug)
	if err != nil {
		return err
	}

	return s.repo.UpdateCategory(ctx, categoryParams)
}

func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	return s.repo.DeleteCategory(ctx, id)
}

func normalize(name, slug string) (string, string, error) {
	name, slug = strings.TrimSpace(name),
		strings.ToLower(strings.TrimSpace(slug))
	switch {
	case name == "":
		return "", "", ErrNameRequired
	case slug == "":
		return "", "", ErrSlugRequired
	}
	return name, slug, nil
}
