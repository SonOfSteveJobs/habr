package tracing

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			propagator := otel.GetTextMapPropagator()
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			ctx, span := otel.Tracer("").Start(ctx,
				fmt.Sprintf("http.server %s %s", r.Method, r.URL.Path),
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
				),
			)
			defer span.End()

			traceID := span.SpanContext().TraceID().String()
			w.Header().Set("X-Trace-Id", traceID)

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r.WithContext(ctx))

			span.SetAttributes(attribute.Int("http.status_code", rw.statusCode))
			if rw.statusCode >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, http.StatusText(rw.statusCode))
			}
		})
	}
}
