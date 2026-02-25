package outbox

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/SonOfSteveJobs/habr/pkg/transaction"
	"github.com/SonOfSteveJobs/habr/services/auth/internal/model"
)

type Repository struct {
	txManager *transaction.Manager
}

func New(txManager *transaction.Manager) *Repository {
	return &Repository{txManager: txManager}
}

func (r *Repository) Insert(ctx context.Context, event model.OutboxEvent) error {
	const query = `
		INSERT INTO outbox (event_id, topic, key, value, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.txManager.ExtractExecutor(ctx).Exec(
		ctx, query,
		event.EventID, event.Topic, event.Key, event.Value, event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("outbox insert: %w", err)
	}

	return nil
}

// FOR UPDATE SKIP LOCKED - блокировка строк + скип заблокированных (возможность запустить несколько relay)
func (r *Repository) FetchUnsent(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	const query = `
		SELECT event_id, topic, key, value, created_at
		FROM outbox
		WHERE NOT is_sent
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.txManager.ExtractExecutor(ctx).Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("outbox fetch unsent: %w", err)
	}
	defer rows.Close()

	var events []model.OutboxEvent
	for rows.Next() {
		var e model.OutboxEvent
		if err := rows.Scan(&e.EventID, &e.Topic, &e.Key, &e.Value, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("outbox scan: %w", err)
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

func (r *Repository) MarkSent(ctx context.Context, eventID uuid.UUID) error {
	const query = `UPDATE outbox SET is_sent = TRUE WHERE event_id = $1`

	_, err := r.txManager.ExtractExecutor(ctx).Exec(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("outbox mark sent: %w", err)
	}

	return nil
}

func (r *Repository) DeleteSent(ctx context.Context) error {
	const query = `DELETE FROM outbox WHERE is_sent`

	_, err := r.txManager.ExtractExecutor(ctx).Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("outbox delete sent: %w", err)
	}

	return nil
}
