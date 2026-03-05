package metrics

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	grpcstatus "google.golang.org/grpc/status"
)

var (
	serverOnce       sync.Once
	serverReqCounter metric.Int64Counter
	serverDuration   metric.Float64Histogram

	clientOnce       sync.Once
	clientReqCounter metric.Int64Counter
	clientDuration   metric.Float64Histogram
)

func initServerMetrics() {
	serverOnce.Do(func() {
		meter := otel.Meter("pkg/metrics")
		serverReqCounter, _ = meter.Int64Counter("rpc.server.request.total",
			metric.WithDescription("Total number of gRPC server requests"),
		)
		serverDuration, _ = meter.Float64Histogram("rpc.server.duration",
			metric.WithDescription("Duration of gRPC server requests in seconds"),
			metric.WithUnit("s"),
		)
	})
}

func initClientMetrics() {
	clientOnce.Do(func() {
		meter := otel.Meter("pkg/metrics")
		clientReqCounter, _ = meter.Int64Counter("rpc.client.request.total",
			metric.WithDescription("Total number of gRPC client requests"),
		)
		clientDuration, _ = meter.Float64Histogram("rpc.client.duration",
			metric.WithDescription("Duration of gRPC client requests in seconds"),
			metric.WithUnit("s"),
		)
	})
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		initServerMetrics()

		start := time.Now()
		resp, err := handler(ctx, req)
		elapsed := time.Since(start).Seconds()

		st, _ := grpcstatus.FromError(err)
		attrs := []attribute.KeyValue{
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.grpc.status_code", st.Code().String()),
		}

		serverReqCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		serverDuration.Record(ctx, elapsed, metric.WithAttributes(
			attribute.String("rpc.method", info.FullMethod),
		))

		return resp, err
	}
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		initClientMetrics()

		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		elapsed := time.Since(start).Seconds()

		st, _ := grpcstatus.FromError(err)
		attrs := []attribute.KeyValue{
			attribute.String("rpc.method", method),
			attribute.String("rpc.grpc.status_code", st.Code().String()),
		}

		clientReqCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		clientDuration.Record(ctx, elapsed, metric.WithAttributes(
			attribute.String("rpc.method", method),
		))

		return err
	}
}
