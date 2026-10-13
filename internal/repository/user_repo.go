package repository

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

type UserRepository struct {
	q *database.Queries
}

func NewUserRepository(q *database.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) Create(ctx context.Context, params database.CreateUserParams) (database.CreateUserRow, error) {
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

func (r *UserRepository) List(ctx context.Context) ([]database.ListUsersRow, error) {
	rows, err := r.q.ListUsers(ctx)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return []database.ListUsersRow{}, pgErr
		}
		return []database.ListUsersRow{}, err
	}

	return rows, nil
}

func (r *UserRepository) Get(ctx context.Context, id int64) (database.GetUserRow, error) {
	row, err := r.q.GetUser(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.GetUserRow{}, ErrUserNotFound
		}
		return database.GetUserRow{}, err
	}

	return row, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (database.GetUserByUsernameRow, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.GetUserByUsernameRow{}, ErrUserNotFound
		}
		return database.GetUserByUsernameRow{}, err
	}

	return row, nil
}
