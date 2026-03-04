package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, MetadataCarrier(md))

		ctx, span := otel.Tracer("").Start(ctx, "grpc.server/"+info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attribute.String("rpc.method", info.FullMethod)),
		)
		defer span.End()

		resp, err := handler(ctx, req)
		if err != nil {
			st, _ := grpcstatus.FromError(err)
			span.SetStatus(codes.Error, st.Message())
			span.SetAttributes(attribute.String("rpc.grpc.status_code", st.Code().String()))
		}

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
		ctx, span := otel.Tracer("").Start(ctx, "grpc.client/"+method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(attribute.String("rpc.method", method)),
		)
		defer span.End()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, MetadataCarrier(md))
		ctx = metadata.NewOutgoingContext(ctx, md)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			st, _ := grpcstatus.FromError(err)
			span.SetStatus(codes.Error, st.Message())
			span.SetAttributes(attribute.String("rpc.grpc.status_code", st.Code().String()))
		}

		return err
	}
}
