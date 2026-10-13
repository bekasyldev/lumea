package user

import (
	"context"

	database "github.com/bekasyldev/lumea/internal/database"
)

type userStore interface {
	Get(ctx context.Context, id int64) (database.GetUserRow, error)
	List(ctx context.Context) ([]database.ListUsersRow, error)
}

type Service struct {
	repo userStore
}

func NewService(repo userStore) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id int64) (database.GetUserRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]database.ListUsersRow, error) {
	return s.repo.List(ctx)
}
