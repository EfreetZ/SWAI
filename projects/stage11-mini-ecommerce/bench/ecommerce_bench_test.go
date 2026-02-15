package bench

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/app"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

func BenchmarkCreateOrderAndPay(b *testing.B) {
	a, _ := app.New()
	p, _ := a.ProductSvc.Create(context.Background(), "p1", 100)
	_ = a.InventorySvc.Seed(context.Background(), p.ID, int64(b.N*2+100))
	items := []model.OrderItem{{ProductID: p.ID, Quantity: 1, Price: 100}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = a.CreateOrderAndPay(context.Background(), 1, items)
	}
}
