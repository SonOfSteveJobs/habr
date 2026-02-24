package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/transaction"
)

type Repository struct {
	txManager *transaction.Manager
}

func New(txManager *transaction.Manager) *Repository {
	return &Repository{txManager: txManager}
}

func (r *Repository) MarkProcessed(ctx context.Context, eventID uuid.UUID) (bool, error) {
	const query = `INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`

	ct, err := r.txManager.ExtractExecutor(ctx).Exec(ctx, query, eventID)
	if err != nil {
		return false, fmt.Errorf("mark processed: %w", err)
	}

	return ct.RowsAffected() == 1, nil
}

func (r *Repository) DeleteOld(ctx context.Context, retention time.Duration) error {
	const query = `DELETE FROM processed_events WHERE processed_at < now() - $1::interval`

	_, err := r.txManager.ExtractExecutor(ctx).Exec(ctx, query, retention.String())
	if err != nil {
		return fmt.Errorf("delete old events: %w", err)
	}

	return nil
}
