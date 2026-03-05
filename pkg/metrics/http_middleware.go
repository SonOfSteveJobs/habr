package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	httpOnce       sync.Once
	httpReqCounter metric.Int64Counter
	httpDuration   metric.Float64Histogram
)

func initHTTPMetrics() {
	httpOnce.Do(func() {
		meter := otel.Meter("pkg/metrics")
		httpReqCounter, _ = meter.Int64Counter("http.server.request.total", //nolint:gosec
			metric.WithDescription("Total number of HTTP server requests"),
		)
		httpDuration, _ = meter.Float64Histogram("http.server.duration", //nolint:gosec
			metric.WithDescription("Duration of HTTP server requests in seconds"),
			metric.WithUnit("s"),
		)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			initHTTPMetrics()

			start := time.Now()
			sw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(sw, r)

			elapsed := time.Since(start).Seconds()

			path := r.URL.Path
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				if pattern := rctx.RoutePattern(); pattern != "" {
					path = pattern
				}
			}

			attrs := []attribute.KeyValue{
				attribute.String("http.method", r.Method),
				attribute.String("http.path", path),
				attribute.String("http.status_code", fmt.Sprintf("%d", sw.statusCode)),
			}

			httpReqCounter.Add(r.Context(), 1, metric.WithAttributes(attrs...))
			httpDuration.Record(r.Context(), elapsed, metric.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.path", path),
			))
		})
	}
}
