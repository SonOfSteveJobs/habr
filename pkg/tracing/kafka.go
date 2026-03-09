package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/SonOfSteveJobs/habr/pkg/kafka"
	"github.com/SonOfSteveJobs/habr/pkg/kafka/consumer"
)

type KafkaHeadersCarrier map[string][]byte

func (c KafkaHeadersCarrier) Get(key string) string {
	if v, ok := c[key]; ok {
		return string(v)
	}

	return ""
}

func (c KafkaHeadersCarrier) Set(key, value string) {
	c[key] = []byte(value)
}

func (c KafkaHeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}

	return keys
}

func InjectToMessage(ctx context.Context, msg *kafka.Message) {
	if msg.Headers == nil {
		msg.Headers = make(map[string][]byte)
	}

	otel.GetTextMapPropagator().Inject(ctx, KafkaHeadersCarrier(msg.Headers))
}

func ExtractFromMessage(ctx context.Context, msg kafka.Message) context.Context {
	if msg.Headers == nil {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, KafkaHeadersCarrier(msg.Headers))
}

func ConsumerMiddleware() consumer.Middleware {
	return func(next kafka.MessageHandler) kafka.MessageHandler {
		return func(ctx context.Context, msg kafka.Message) error {
			ctx = ExtractFromMessage(ctx, msg)

			ctx, span := otel.Tracer("").Start(ctx, "kafka.consume/"+msg.Topic,
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(
					attribute.String("messaging.system", "kafka"),
					attribute.String("messaging.destination", msg.Topic),
					attribute.Int64("messaging.kafka.offset", msg.Offset),
					attribute.Int("messaging.kafka.partition", int(msg.Partition)),
				),
			)
			defer span.End()

			err := next(ctx, msg)
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}

			return err
		}
	}
}
