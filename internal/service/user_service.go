package service

import (
	"context"

	database "github.com/bekasyldev/lumea/internal/database"
)

type UserRepository interface {
	Create(ctx context.Context, params database.CreateUserParams) (database.CreateUserRow, error)
	Get(ctx context.Context, id int64) (database.GetUserRow, error)
	List(ctx context.Context) ([]database.ListUsersRow, error)
	GetByUsername(ctx context.Context, username string) (database.GetUserByUsernameRow, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetById(ctx context.Context, id int64) (database.GetUserRow, error) {
	return s.repo.Get(ctx, id)
}

func (s *UserService) GetList(ctx context.Context) ([]database.ListUsersRow, error) {
	return s.repo.List(ctx)
}
