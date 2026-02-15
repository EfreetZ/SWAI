package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/app"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

func TestCreateOrderAPI(t *testing.T) {
	a, err := app.New()
	if err != nil {
		t.Fatalf("new app failed: %v", err)
	}
	p, _ := a.ProductSvc.Create(context.Background(), "p1", 100)
	_ = a.InventorySvc.Seed(context.Background(), p.ID, 10)

	s := New(a)
	reqBody := map[string]interface{}{"user_id": 1, "items": []model.OrderItem{{ProductID: p.ID, Quantity: 1, Price: 100}}}
	buf, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewReader(buf))
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", w.Code)
	}
}
