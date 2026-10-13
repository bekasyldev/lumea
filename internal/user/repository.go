package user

import (
	"context"
	"errors"

	database "github.com/bekasyldev/lumea/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Repository struct {
	q *database.Queries
}

func NewRepository(q *database.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) List(ctx context.Context) ([]database.ListUsersRow, error) {
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

func (r *Repository) Get(ctx context.Context, id int64) (database.GetUserRow, error) {
	row, err := r.q.GetUser(ctx, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.GetUserRow{}, ErrUserNotFound
		}
		return database.GetUserRow{}, err
	}

	return row, nil
}
