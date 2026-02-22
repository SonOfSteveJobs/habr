package token

import (
	"context"
	"crypto/sha256"
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

func (r *Repository) Save(ctx context.Context, refreshToken string, userID uuid.UUID, ttl time.Duration) error {
	return r.client.Set(ctx, userID.String(), hashToken(refreshToken), ttl).Err()
}

func (r *Repository) Validate(ctx context.Context, refreshToken string, userID uuid.UUID) error {
	stored, err := r.client.Get(ctx, userID.String()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.ErrInvalidRefreshToken
		}

		return err
	}

	if stored != hashToken(refreshToken) {
		return model.ErrInvalidRefreshToken
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.client.Del(ctx, userID.String()).Err()
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return fmt.Sprintf("%x", hash)
}
