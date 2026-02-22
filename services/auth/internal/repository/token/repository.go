package token

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return fmt.Sprintf("%x", hash)
}
