package logger

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type OTelConfig interface {
	Endpoint() string
	ServiceName() string
	Environment() string
	Version() string
}

var lp *sdklog.LoggerProvider

func InitOTelLogger(ctx context.Context, cfg OTelConfig) error {
	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(cfg.Endpoint()),
		otlploggrpc.WithInsecure(),
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

	lp = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	global.SetLoggerProvider(lp)

	return nil
}

func ShutdownOTelLogger(ctx context.Context) error {
	if lp == nil {
		return nil
	}

	return lp.Shutdown(ctx)
}
