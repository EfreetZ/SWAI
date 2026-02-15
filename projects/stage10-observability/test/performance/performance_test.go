package performance

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/server"
)

func TestPerformance(t *testing.T) {
	s := server.NewServer()
	h := s.Handler()
	start := time.Now()
	for i := 0; i < 5000; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", w.Code)
		}
	}
	if time.Since(start) > 3*time.Second {
		t.Fatalf("performance regression: %s", time.Since(start))
	}
}
