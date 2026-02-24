package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/SonOfSteveJobs/habr/pkg/transaction"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type Repository struct {
	txManager *transaction.Manager
}

func New(txManager *transaction.Manager) *Repository {
	return &Repository{txManager: txManager}
}

func (r *Repository) Create(ctx context.Context, user *model.User) error {
	const query = `
		INSERT INTO users (id, email, hashed_password)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
	`

	ct, err := r.txManager.ExtractExecutor(ctx).Exec(ctx, query, user.ID, user.Email, user.HashedPassword)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return model.ErrEmailAlreadyExists
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

	err := r.txManager.ExtractExecutor(ctx).QueryRow(ctx, query, email).Scan(
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
