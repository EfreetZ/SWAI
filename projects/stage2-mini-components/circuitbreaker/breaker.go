package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrOpenState = errors.New("circuit breaker is open")

// State 是熔断器状态。
type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

// Breaker 三态熔断器。
type Breaker struct {
	mu sync.Mutex

	state            State
	failThreshold    int
	successThreshold int
	timeout          time.Duration
	halfOpenMax      int

	failureCount int
	successCount int
	lastFailTime time.Time
	halfOpenCurr int
}

// New 创建熔断器。
func New(failThreshold, successThreshold int, timeout time.Duration, halfOpenMax int) *Breaker {
	if failThreshold <= 0 {
		failThreshold = 3
	}
	if successThreshold <= 0 {
		successThreshold = 2
	}
	if timeout <= 0 {
		timeout = time.Second
	}
	if halfOpenMax <= 0 {
		halfOpenMax = 1
	}

	return &Breaker{
		state:            Closed,
		failThreshold:    failThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		halfOpenMax:      halfOpenMax,
	}
}

// Execute 执行请求并驱动状态迁移。
func (b *Breaker) Execute(fn func() error) error {
	if fn == nil {
		return errors.New("fn must not be nil")
	}

	if !b.allowRequest() {
		return ErrOpenState
	}

	err := fn()
	b.recordResult(err)
	return err
}

// State 返回当前状态。
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Reset 重置熔断器。
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = Closed
	b.failureCount = 0
	b.successCount = 0
	b.halfOpenCurr = 0
}

func (b *Breaker) allowRequest() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case Closed:
		return true
	case Open:
		if time.Since(b.lastFailTime) >= b.timeout {
			b.state = HalfOpen
			b.halfOpenCurr = 0
			b.successCount = 0
			return true
		}
		return false
	case HalfOpen:
		if b.halfOpenCurr >= b.halfOpenMax {
			return false
		}
		b.halfOpenCurr++
		return true
	default:
		return false
	}
}

func (b *Breaker) recordResult(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err == nil {
		switch b.state {
		case Closed:
			b.failureCount = 0
		case HalfOpen:
			b.successCount++
			if b.successCount >= b.successThreshold {
				b.state = Closed
				b.failureCount = 0
				b.successCount = 0
				b.halfOpenCurr = 0
			}
		}
		return
	}

	switch b.state {
	case Closed:
		b.failureCount++
		if b.failureCount >= b.failThreshold {
			b.state = Open
			b.lastFailTime = time.Now()
		}
	case HalfOpen:
		b.state = Open
		b.lastFailTime = time.Now()
		b.successCount = 0
		b.halfOpenCurr = 0
	}
}
