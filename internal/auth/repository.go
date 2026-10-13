package auth

import (
	"context"
	"errors"

	database "github.com/bekasyldev/lumea/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUsernameTaken = errors.New("username already taken")
	ErrUserNotFound  = errors.New("user not found")
)

type Repository struct {
	q *database.Queries
}

func NewRepository(q *database.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) Create(ctx context.Context, params database.CreateUserParams) (database.CreateUserRow, error) {
	row, err := r.q.CreateUser(ctx, database.CreateUserParams{
		Username:       params.Username,
		HashedPassword: params.HashedPassword,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return database.CreateUserRow{}, ErrUsernameTaken
		}
		return database.CreateUserRow{}, err
	}
	return row, nil
}

func (r *Repository) Get(ctx context.Context, username string) (database.GetUserByUsernameRow, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.GetUserByUsernameRow{}, ErrUserNotFound
		}
		return database.GetUserByUsernameRow{}, err
	}

	return row, nil
}
