package product

import (
	"context"
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

// Service 商品服务。
type Service struct {
	mu       sync.RWMutex
	products map[int64]model.Product
	nextID   int64
}

func NewService() *Service {
	return &Service{products: make(map[int64]model.Product), nextID: 1}
}

func (s *Service) Create(ctx context.Context, name string, price int64) (model.Product, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.Product{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p := model.Product{ID: s.nextID, Name: name, Price: price}
	s.nextID++
	s.products[p.ID] = p
	return p, nil
}

func (s *Service) Get(ctx context.Context, id int64) (model.Product, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.Product{}, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	return p, ok, nil
}
