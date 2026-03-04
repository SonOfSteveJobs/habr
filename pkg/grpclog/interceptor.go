package grpclog

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/SonOfSteveJobs/habr/pkg/logger"
)

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		log := logger.Ctx(ctx)

		if err != nil {
			st, _ := status.FromError(err)
			log.Error().
				Str("grpc_method", method).
				Str("grpc_code", st.Code().String()).
				Str("grpc_message", st.Message()).
				Dur("duration", duration).
				Msg("grpc client call")
		} else {
			log.Info().
				Str("grpc_method", method).
				Dur("duration", duration).
				Msg("grpc client call")
		}

		return err
	}
}
