package verification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type Repository struct {
	client *redis.Client
}

func New(client *redis.Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) Save(ctx context.Context, code string, userID uuid.UUID, ttl time.Duration) error {
	return r.client.Set(ctx, key(userID), code, ttl).Err()
}

func (r *Repository) Validate(ctx context.Context, code string, userID uuid.UUID) error {
	stored, err := r.client.Get(ctx, key(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.ErrInvalidVerificationCode
		}

		return err
	}

	if stored != code {
		return model.ErrInvalidVerificationCode
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.client.Del(ctx, key(userID)).Err()
}

func key(userID uuid.UUID) string {
	return fmt.Sprintf("verify:%s", userID.String())
}
