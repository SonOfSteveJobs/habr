package transaction

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/SonOfSteveJobs/habr/pkg/metrics"
)

type instrumentedExecutor struct {
	inner Executor
}

func newInstrumented(e Executor) Executor {
	return &instrumentedExecutor{inner: e}
}

func (i *instrumentedExecutor) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	ct, err := i.inner.Exec(ctx, sql, args...)
	metrics.RecordDBOperation(ctx, "exec", start)
	return ct, err
}

func (i *instrumentedExecutor) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	start := time.Now()
	rows, err := i.inner.Query(ctx, sql, args...)
	metrics.RecordDBOperation(ctx, "query", start)
	return rows, err
}

func (i *instrumentedExecutor) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	start := time.Now()
	row := i.inner.QueryRow(ctx, sql, args...)
	metrics.RecordDBOperation(ctx, "query_row", start)
	return row
}
