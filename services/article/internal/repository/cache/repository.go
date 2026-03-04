package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

const cacheKey = "articles:first_page"

type cachedPage struct {
	Articles   []cachedArticle `json:"articles"`
	NextCursor string          `json:"next_cursor"`
}

type cachedArticle struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, ttl time.Duration) *Repository {
	return &Repository{client: client, ttl: ttl}
}

func (r *Repository) Get(ctx context.Context) (*model.ArticlePage, error) {
	data, err := r.client.Get(ctx, cacheKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var cached cachedPage
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("unmarshal cached page: %w", err)
	}

	articles := make([]*model.Article, len(cached.Articles))
	for i, a := range cached.Articles {
		articles[i] = &model.Article{
			ID:        a.ID,
			AuthorID:  a.AuthorID,
			Title:     a.Title,
			Content:   a.Content,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		}
	}

	return &model.ArticlePage{
		Articles:   articles,
		NextCursor: cached.NextCursor,
	}, nil
}

func (r *Repository) Set(ctx context.Context, page *model.ArticlePage) error {
	cached := cachedPage{
		Articles:   make([]cachedArticle, len(page.Articles)),
		NextCursor: page.NextCursor,
	}

	for i, a := range page.Articles {
		cached.Articles[i] = cachedArticle{
			ID:        a.ID,
			AuthorID:  a.AuthorID,
			Title:     a.Title,
			Content:   a.Content,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		}
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("marshal cached page: %w", err)
	}

	return r.client.Set(ctx, cacheKey, data, r.ttl).Err()
}

func (r *Repository) Invalidate(ctx context.Context) error {
	return r.client.Del(ctx, cacheKey).Err()
}
