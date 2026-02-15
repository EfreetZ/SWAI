package resilience

import (
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker open")

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker 熔断器。
type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failureCount int64
	threshold    int64
	timeout      time.Duration
	lastFailTime time.Time
}

// NewCircuitBreaker 创建熔断器。
func NewCircuitBreaker(threshold int64, timeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if timeout <= 0 {
		timeout = time.Second
	}
	return &CircuitBreaker{state: StateClosed, threshold: threshold, timeout: timeout}
}

// Execute 执行受保护调用。
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.AllowRequest() {
		return ErrCircuitOpen
	}
	err := fn()
	if err != nil {
		cb.RecordFailure()
		return err
	}
	cb.RecordSuccess()
	return nil
}

// AllowRequest 判断是否允许请求通过。
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) >= cb.timeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	}
	return true
}

// RecordFailure 记录失败。
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	cb.lastFailTime = time.Now()
	if cb.failureCount >= cb.threshold {
		cb.state = StateOpen
	}
}

// RecordSuccess 记录成功。
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = StateClosed
}

// State 获取当前状态。
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
