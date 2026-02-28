package article

import (
	"context"

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
