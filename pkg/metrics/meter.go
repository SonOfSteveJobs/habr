package metrics

import (
	"context"

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

	mp = metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(res),
	)

	otel.SetMeterProvider(mp)

	return nil
}

func ShutdownMeter(ctx context.Context) error {
	if mp == nil {
		return nil
	}

	return mp.Shutdown(ctx)
}
