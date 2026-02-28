package article

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/SonOfSteveJobs/habr/pkg/transaction"
	"github.com/SonOfSteveJobs/habr/services/article/internal/model"
)

type Repository struct {
	txManager *transaction.Manager
}

func New(txManager *transaction.Manager) *Repository {
	return &Repository{txManager: txManager}
}

func (r *Repository) Create(ctx context.Context, article *model.Article) error {
	const query = `
		INSERT INTO articles (id, author_id, title, content)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	return r.txManager.ExtractExecutor(ctx).QueryRow(
		ctx, query,
		article.ID, article.AuthorID, article.Title, article.Content,
	).Scan(&article.CreatedAt, &article.UpdatedAt)
}

func (r *Repository) List(ctx context.Context, cursor string, limit int) (*model.ArticlePage, error) {
	var (
		rows pgx.Rows
		err  error
	)

	exec := r.txManager.ExtractExecutor(ctx)

	if cursor == "" {
		const query = `
			SELECT id, author_id, title, content, created_at, updated_at
			FROM articles
			ORDER BY created_at DESC, id DESC
			LIMIT $1
		`
		rows, err = exec.Query(ctx, query, limit+1)
	} else {
		createdAt, id, parseErr := model.DecodeCursor(cursor)
		if parseErr != nil {
			return nil, fmt.Errorf("decode cursor: %w", parseErr)
		}

		const query = `
			SELECT id, author_id, title, content, created_at, updated_at
			FROM articles
			WHERE (created_at, id) < ($1, $2)
			ORDER BY created_at DESC, id DESC
			LIMIT $3
		`
		rows, err = exec.Query(ctx, query, createdAt, id, limit+1)
	}

	if err != nil {
		return nil, fmt.Errorf("query articles: %w", err)
	}
	defer rows.Close()

	articles, err := scanArticles(rows, limit)
	if err != nil {
		return nil, err
	}

	page := &model.ArticlePage{Articles: articles}

	if len(articles) > limit {
		page.Articles = articles[:limit]
		last := page.Articles[limit-1]
		page.NextCursor = model.EncodeCursor(last.CreatedAt, last.ID)
	}

	return page, nil
}

func scanArticles(rows pgx.Rows, capacity int) ([]*model.Article, error) {
	articles := make([]*model.Article, 0, capacity+1)

	for rows.Next() {
		var a model.Article

		if err := rows.Scan(&a.ID, &a.AuthorID, &a.Title, &a.Content, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}

		articles = append(articles, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return articles, nil
}
