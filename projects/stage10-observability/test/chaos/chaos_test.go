package chaos

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/metrics"
)

func TestMetricsHandlerChaos(t *testing.T) {
	m := &metrics.REDMetrics{}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.IncRequest(100, true)
		h.ServeHTTP(w, r)
	})
	wrapped.ServeHTTP(mw, req)
	if m.ErrorRate() <= 0 {
		t.Fatal("expected error rate > 0")
	}
}
