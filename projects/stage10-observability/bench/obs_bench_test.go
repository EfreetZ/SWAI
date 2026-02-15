package bench

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/server"
)

func BenchmarkPing(b *testing.B) {
	s := server.NewServer()
	h := s.Handler()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}
}
