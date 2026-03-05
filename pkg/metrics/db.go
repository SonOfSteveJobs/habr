package metrics

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	dbOnce     sync.Once
	dbCounter  metric.Int64Counter
	dbDuration metric.Float64Histogram
)

func initDBMetrics() {
	dbOnce.Do(func() {
		meter := otel.Meter("pkg/metrics")
		dbCounter, _ = meter.Int64Counter("db.client.operation.total",
			metric.WithDescription("Total number of database operations"),
		)
		dbDuration, _ = meter.Float64Histogram("db.client.operation.duration",
			metric.WithDescription("Duration of database operations in seconds"),
			metric.WithUnit("s"),
		)
	})
}

func RecordDBOperation(ctx context.Context, operation string, start time.Time) {
	initDBMetrics()

	elapsed := time.Since(start).Seconds()
	attrs := metric.WithAttributes(
		attribute.String("db.operation", operation),
	)

	dbCounter.Add(ctx, 1, attrs)
	dbDuration.Record(ctx, elapsed, attrs)
}
