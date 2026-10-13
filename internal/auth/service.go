package auth

import (
	"context"
	"errors"

	database "github.com/bekasyldev/lumea/internal/database"
	token "github.com/bekasyldev/lumea/internal/token"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type authStore interface {
	Create(ctx context.Context, params database.CreateUserParams) (database.CreateUserRow, error)
	Get(ctx context.Context, username string) (database.GetUserByUsernameRow, error)
}

type Service struct {
	repo authStore
}

func NewService(repo authStore) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, username, password string) (database.CreateUserRow, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return database.CreateUserRow{}, err
	}

	return s.repo.Create(ctx, database.CreateUserParams{
		Username:       username,
		HashedPassword: string(hashed),
	})
}

func (s *Service) Login(ctx context.Context, username, password string) (token.TokenPair, error) {
	user, err := s.repo.Get(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return token.TokenPair{}, ErrInvalidCredentials
		}
		return token.TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return token.TokenPair{}, ErrInvalidCredentials
	}

	accessToken, err := token.IssueAccessToken(user.Username, string(user.Role))
	if err != nil {
		return token.TokenPair{}, err
	}

	refreshToken, err := token.IssueRefreshToken(user.Username)
	if err != nil {
		return token.TokenPair{}, err
	}

	return token.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
