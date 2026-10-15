package categories

import (
	"context"
	"errors"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrSlugTaken        = errors.New("slug should be unique")
	ErrCategoryNotFound = errors.New("category did not exist")
)

type Repository struct {
	q *database.Queries
}

func NewRepository(q *database.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) ListCategories(ctx context.Context) ([]database.Category, error) {
	categories, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	// encode an empty list as [] instead of null
	if categories == nil {
		return []database.Category{}, nil
	}

	return categories, nil
}

func (r *Repository) GetCategoryByID(ctx context.Context, id int64) (database.Category, error) {
	category, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.Category{}, ErrCategoryNotFound
		}
		return database.Category{}, err
	}

	return category, nil
}

func (r *Repository) GetCategoryBySlug(ctx context.Context, slug string) (database.Category, error) {
	category, err := r.q.GetCategoryBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.Category{}, ErrCategoryNotFound
		}
		return database.Category{}, err
	}

	return category, nil
}

func (r *Repository) CreateCategory(ctx context.Context, categoryParams database.CreateCategoryParams) (database.Category, error) {
	category, err := r.q.CreateCategory(ctx, categoryParams)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505 error code for unique violation
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return database.Category{}, ErrSlugTaken
		}
		return database.Category{}, err
	}

	return category, nil
}

func (r *Repository) UpdateCategory(ctx context.Context, categoryParams database.UpdateCategoryParams) error {
	rows, err := r.q.UpdateCategory(ctx, categoryParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrSlugTaken
		}
		return err
	}
	if rows == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id int64) error {
	rows, err := r.q.DeleteCategory(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCategoryNotFound
	}

	return nil
}
