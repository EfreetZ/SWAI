package breaker

import (
	"errors"
	"sync"
	"time"
)

var ErrOpen = errors.New("circuit open")

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

// CircuitBreaker 熔断器。
type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failureCount int64
	threshold    int64
	timeout      time.Duration
	lastFailure  time.Time
}

// New 创建熔断器。
func New(threshold int64, timeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if timeout <= 0 {
		timeout = time.Second
	}
	return &CircuitBreaker{state: Closed, threshold: threshold, timeout: timeout}
}

// Execute 执行受保护请求。
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allow() {
		return ErrOpen
	}
	err := fn()
	if err != nil {
		cb.onFailure()
		return err
	}
	cb.onSuccess()
	return nil
}

func (cb *CircuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == Open {
		if time.Since(cb.lastFailure) >= cb.timeout {
			cb.state = HalfOpen
			return true
		}
		return false
	}
	return true
}

func (cb *CircuitBreaker) onFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	cb.lastFailure = time.Now()
	if cb.failureCount >= cb.threshold {
		cb.state = Open
	}
}

func (cb *CircuitBreaker) onSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = Closed
}
