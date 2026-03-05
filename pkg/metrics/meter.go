package metrics

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Config interface {
	Endpoint() string
	ServiceName() string
	Environment() string
	Version() string
}

var mp *metric.MeterProvider

func InitMeter(ctx context.Context, cfg Config) error {
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint()),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName()),
			semconv.ServiceVersionKey.String(cfg.Version()),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment()),
		),
	)
	if err != nil {
		return err
	}

	// Use explicit bucket histograms so Prometheus stores classic _bucket series
	// that dashboard PromQL queries (histogram_quantile) can work with.
	defaultBuckets := []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	histogramView := metric.NewView(
		metric.Instrument{Kind: metric.InstrumentKindHistogram},
		metric.Stream{
			Aggregation: metric.AggregationExplicitBucketHistogram{
				Boundaries: defaultBuckets,
			},
		},
	)

	mp = metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
		metric.WithView(histogramView),
	)

	otel.SetMeterProvider(mp)

	if err := runtime.Start(); err != nil {
		return err
	}

	return nil
}

func ShutdownMeter(ctx context.Context) error {
	if mp == nil {
		return nil
	}

	return mp.Shutdown(ctx)
}
