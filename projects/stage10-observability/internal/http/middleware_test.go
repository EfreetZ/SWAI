package httpobs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/metrics"
)

func TestMetricsMiddleware(t *testing.T) {
	m := &metrics.REDMetrics{}
	h := MetricsMiddleware(m, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
	if w.Header().Get("X-Trace-ID") == "" {
		t.Fatal("trace id missing")
	}
}
