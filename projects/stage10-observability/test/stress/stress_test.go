package stress

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/server"
)

func TestStress(t *testing.T) {
	s := server.NewServer()
	h := s.Handler()
	const workers = 20
	const each = 200
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < each; j++ {
				req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
				w := httptest.NewRecorder()
				h.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					t.Errorf("unexpected status: %d", w.Code)
					return
				}
			}
		}()
	}
	wg.Wait()
}
