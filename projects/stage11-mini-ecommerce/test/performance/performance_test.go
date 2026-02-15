package performance

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/app"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

func TestPerformance(t *testing.T) {
	a, _ := app.New()
	p, _ := a.ProductSvc.Create(context.Background(), "p1", 100)
	_ = a.InventorySvc.Seed(context.Background(), p.ID, 10000)
	items := []model.OrderItem{{ProductID: p.ID, Quantity: 1, Price: 100}}
	start := time.Now()
	for i := 0; i < 3000; i++ {
		if _, _, err := a.CreateOrderAndPay(context.Background(), 1, items); err != nil {
			t.Fatalf("flow failed: %v", err)
		}
	}
	if time.Since(start) > 3*time.Second {
		t.Fatalf("performance regression: %s", time.Since(start))
	}
}
