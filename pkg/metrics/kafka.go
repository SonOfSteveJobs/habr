package metrics

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/kafka/consumer"
)

var (
	kafkaOnce             sync.Once
	kafkaConsumerCounter  metric.Int64Counter
	kafkaConsumerDuration metric.Float64Histogram
	kafkaProducerCounter  metric.Int64Counter
)

func initKafkaMetrics() {
	kafkaOnce.Do(func() {
		meter := otel.Meter("pkg/metrics")
		kafkaConsumerCounter, _ = meter.Int64Counter("kafka.consumer.messages.total",
			metric.WithDescription("Total number of consumed Kafka messages"),
		)
		kafkaConsumerDuration, _ = meter.Float64Histogram("kafka.consumer.duration",
			metric.WithDescription("Duration of Kafka message processing in seconds"),
			metric.WithUnit("s"),
		)
		kafkaProducerCounter, _ = meter.Int64Counter("kafka.producer.messages.total",
			metric.WithDescription("Total number of produced Kafka messages"),
		)
	})
}

func ConsumerMiddleware() consumer.Middleware {
	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) error {
			initKafkaMetrics()

			start := time.Now()
			err := next(ctx, msg)
			elapsed := time.Since(start).Seconds()

			attrs := attribute.String("messaging.destination", msg.Topic)

			kafkaConsumerCounter.Add(ctx, 1, metric.WithAttributes(attrs))
			kafkaConsumerDuration.Record(ctx, elapsed, metric.WithAttributes(attrs))

			return err
		}
	}
}

func RecordProducerMessage(ctx context.Context, topic string) {
	initKafkaMetrics()

	kafkaProducerCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("messaging.destination", topic),
	))
}
