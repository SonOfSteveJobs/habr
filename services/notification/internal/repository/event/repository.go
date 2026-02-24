package event

import (
	"context"
	"fmt"

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

func (r *Repository) DeleteOld(ctx context.Context) error {
	const query = `DELETE FROM processed_events WHERE processed_at < now() - INTERVAL '7 days'`

	_, err := r.txManager.ExtractExecutor(ctx).Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("delete old events: %w", err)
	}

	return nil
}
