package payment

import (
	"context"
	"sync"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/infra"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

// Service 支付服务。
type Service struct {
	mu       sync.RWMutex
	payments map[string]model.Payment
	idGen    *infra.Snowflake
}

func NewService(idGen *infra.Snowflake) *Service {
	return &Service{payments: make(map[string]model.Payment), idGen: idGen}
}

func (s *Service) Create(ctx context.Context, orderID string, amount int64) (model.Payment, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.Payment{}, err
	}
	id, err := s.idGen.NextID()
	if err != nil {
		return model.Payment{}, err
	}
	p := model.Payment{ID: formatID(id), OrderID: orderID, Amount: amount, Success: false, CreatedAt: time.Now()}
	s.mu.Lock()
	s.payments[p.ID] = p
	s.mu.Unlock()
	return p, nil
}

func (s *Service) Callback(ctx context.Context, paymentID string, success bool) (model.Payment, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.Payment{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.payments[paymentID]
	if !ok {
		return model.Payment{}, false, nil
	}
	p.Success = success
	s.payments[paymentID] = p
	return p, true, nil
}

func formatID(v int64) string {
	return "pay-" + int64ToString(v)
}

func int64ToString(v int64) string {
	if v == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	return string(buf)
}
