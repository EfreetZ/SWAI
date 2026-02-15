package stress

import (
	"context"
	"sync"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/app"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

func TestStress(t *testing.T) {
	a, _ := app.New()
	p, _ := a.ProductSvc.Create(context.Background(), "p1", 100)
	_ = a.InventorySvc.Seed(context.Background(), p.ID, 50000)
	items := []model.OrderItem{{ProductID: p.ID, Quantity: 1, Price: 100}}

	const workers = 20
	const each = 200
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < each; j++ {
				_, _, err := a.CreateOrderAndPay(context.Background(), 1, items)
				if err != nil {
					t.Errorf("flow failed: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
