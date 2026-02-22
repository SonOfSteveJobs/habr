package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

const uniqueViolationCode = "23505"

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, user *model.User) error {
	const query = `
		INSERT INTO users (id, email, hashed_password)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.HashedPassword)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolationCode {
			return model.ErrEmailAlreadyExists
		}

		return err
	}

	return nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const query = `
		SELECT id, email, hashed_password, is_email_confirmed, created_at
		FROM users
		WHERE email = $1
	`

	var user model.User

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.HashedPassword,
		&user.IsEmailConfirmed,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}
