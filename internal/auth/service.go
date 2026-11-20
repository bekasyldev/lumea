package auth

import (
	"context"
	"errors"
	"fmt"

	database "github.com/bekasyldev/lumea/internal/database"
	"github.com/bekasyldev/lumea/internal/store"
	token "github.com/bekasyldev/lumea/internal/token"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
var ErrPersistToken = errors.New("could not persist new tokens")

type authStore interface {
	Create(ctx context.Context, params database.CreateUserParams) (database.CreateUserRow, error)
	Get(ctx context.Context, username string) (database.GetUserByUsernameRow, error)
}

type Service struct {
	redis *store.Redis
	repo  authStore
}

func NewService(redis *store.Redis, repo authStore) *Service {
	return &Service{redis: redis, repo: repo}
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

	claims := &token.Claims{}
	_, err = token.ValidateToken(refreshToken, token.TokenRefresh, claims)
	if err != nil {
		return token.TokenPair{}, ErrInvalidRefreshToken
	}

	if err := token.Persist(ctx, s.redis, *claims); err != nil {
		return token.TokenPair{}, ErrPersistToken
	}

	return token.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (token.TokenPair, error) {
	oldClaims := &token.Claims{}
	parsed, err := token.ValidateToken(refreshToken, token.TokenRefresh, oldClaims)
	if err != nil || !parsed.Valid || oldClaims.Role != "" || oldClaims.Subject == "" {
		return token.TokenPair{}, ErrInvalidRefreshToken
	}

	key := "refresh:" + oldClaims.ID
	if _, err := s.redis.ConsumeJTI(ctx, key); err != nil {
		return token.TokenPair{}, ErrInvalidRefreshToken
	}

	user, err := s.repo.Get(ctx, oldClaims.Subject)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return token.TokenPair{}, ErrInvalidRefreshToken
		}
		return token.TokenPair{}, err
	}

	access, err := token.IssueAccessToken(user.Username, string(user.Role))
	if err != nil {
		return token.TokenPair{}, err
	}
	refresh, err := token.IssueRefreshToken(user.Username)
	if err != nil {
		return token.TokenPair{}, err
	}

	newClaims := &token.Claims{}
	_, err = token.ValidateToken(refresh, token.TokenRefresh, newClaims)
	if err != nil {
		return token.TokenPair{}, ErrInvalidRefreshToken
	}

	if err := token.Persist(ctx, s.redis, *newClaims); err != nil {
		return token.TokenPair{}, ErrPersistToken
	}

	return token.TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims := &token.Claims{}
	parsed, err := token.ValidateToken(refreshToken, token.TokenRefresh, claims)
	if err != nil || !parsed.Valid {
		return ErrInvalidRefreshToken
	}

	if _, err := s.redis.ConsumeJTI(ctx, "refresh:"+claims.ID); err != nil {
		return fmt.Errorf("delete refresh jti: %w", err)
	}

	return nil
}
