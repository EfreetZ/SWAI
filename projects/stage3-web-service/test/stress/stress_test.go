package stress

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	stage3test "github.com/EfreetZ/SWAI/projects/stage3-web-service/test"
)

// TestRegisterStress 并发注册压力测试。
func TestRegisterStress(t *testing.T) {
	app, _ := stage3test.NewTestAppForExternal()

	const workers = 40
	const each = 20
	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < each; i++ {
				username := "u" + strconv.Itoa(workerID) + "-" + strconv.Itoa(i)
				body := map[string]string{"username": username, "email": username + "@example.com", "password": "123456"}
				payload, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(payload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					t.Errorf("register status = %d", w.Code)
				}
			}
		}(w)
	}
	wg.Wait()
}
