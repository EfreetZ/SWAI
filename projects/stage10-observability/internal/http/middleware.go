package httpobs

import (
	"net/http"
	"strconv"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/metrics"
	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/tracing"
)

// MetricsMiddleware 记录 RED 指标与 trace header。
func MetricsMiddleware(red *metrics.REDMetrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracing.StartSpan(r.Context())
		r = r.WithContext(ctx)
		start := time.Now()
		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		durationMicros := time.Since(start).Microseconds()
		red.IncRequest(durationMicros, rw.status >= 500)
		w.Header().Set("X-Trace-ID", span.TraceID)
		w.Header().Set("X-Span-ID", span.SpanID)
	})
}

// MetricsHandler 暴露文本指标。
func MetricsHandler(red *metrics.REDMetrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, errCnt, avg := red.Snapshot()
		_, _ = w.Write([]byte("requests_total " + strconv.FormatInt(req, 10) + "\n"))
		_, _ = w.Write([]byte("errors_total " + strconv.FormatInt(errCnt, 10) + "\n"))
		_, _ = w.Write([]byte("avg_duration_micros " + strconv.FormatFloat(avg, 'f', 2, 64) + "\n"))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
