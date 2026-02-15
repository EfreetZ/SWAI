package balancer

import (
	"errors"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/config"
)

var ErrNoTarget = errors.New("no target")

// Balancer 负载均衡。
type Balancer interface {
	Pick(targets []config.TargetConfig) (config.TargetConfig, error)
}

// RoundRobin 轮询。
type RoundRobin struct {
	counter uint64
}

func (b *RoundRobin) Pick(targets []config.TargetConfig) (config.TargetConfig, error) {
	if len(targets) == 0 {
		return config.TargetConfig{}, ErrNoTarget
	}
	idx := atomic.AddUint64(&b.counter, 1)
	return targets[idx%uint64(len(targets))], nil
}

// Random 随机。
type Random struct {
	rnd *rand.Rand
}

func NewRandom() *Random {
	return &Random{rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (b *Random) Pick(targets []config.TargetConfig) (config.TargetConfig, error) {
	if len(targets) == 0 {
		return config.TargetConfig{}, ErrNoTarget
	}
	return targets[b.rnd.Intn(len(targets))], nil
}
