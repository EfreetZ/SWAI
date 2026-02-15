package chaos

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/app"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

func TestInsufficientStockChaos(t *testing.T) {
	a, _ := app.New()
	p, _ := a.ProductSvc.Create(context.Background(), "p1", 100)
	_ = a.InventorySvc.Seed(context.Background(), p.ID, 1)
	items := []model.OrderItem{{ProductID: p.ID, Quantity: 2, Price: 100}}
	if _, _, err := a.CreateOrderAndPay(context.Background(), 1, items); err == nil {
		t.Fatal("expected insufficient stock error")
	}
}
