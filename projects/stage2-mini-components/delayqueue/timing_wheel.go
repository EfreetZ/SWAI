package delayqueue

import (
	"context"
	"sync"
	"time"
)

type scheduledTask struct {
	ticksRemaining int
	callback       func()
}

// TimingWheel 是单层时间轮实现。
type TimingWheel struct {
	tickDuration time.Duration
	wheelSize    int
	slots        [][]scheduledTask
	currentPos   int

	mu sync.Mutex
}

// NewTimingWheel 创建时间轮。
func NewTimingWheel(tickDuration time.Duration, wheelSize int) *TimingWheel {
	if tickDuration <= 0 {
		tickDuration = 100 * time.Millisecond
	}
	if wheelSize <= 0 {
		wheelSize = 64
	}
	slots := make([][]scheduledTask, wheelSize)
	return &TimingWheel{tickDuration: tickDuration, wheelSize: wheelSize, slots: slots}
}

// Add 添加延迟任务。
func (tw *TimingWheel) Add(delay time.Duration, callback func()) {
	if callback == nil {
		return
	}
	if delay < 0 {
		delay = 0
	}

	ticks := int(delay / tw.tickDuration)
	if delay%tw.tickDuration != 0 {
		ticks++
	}
	if ticks <= 0 {
		ticks = 1
	}

	tw.mu.Lock()
	defer tw.mu.Unlock()
	pos := (tw.currentPos + ticks) % tw.wheelSize
	rounds := (ticks - 1) / tw.wheelSize
	tw.slots[pos] = append(tw.slots[pos], scheduledTask{ticksRemaining: rounds, callback: callback})
}

// Start 启动时间轮。
func (tw *TimingWheel) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	ticker := time.NewTicker(tw.tickDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tw.tick()
		}
	}
}

func (tw *TimingWheel) tick() {
	tw.mu.Lock()
	tw.currentPos = (tw.currentPos + 1) % tw.wheelSize
	slot := tw.slots[tw.currentPos]
	remaining := make([]scheduledTask, 0, len(slot))
	runnable := make([]func(), 0, len(slot))
	for _, task := range slot {
		if task.ticksRemaining > 0 {
			task.ticksRemaining--
			remaining = append(remaining, task)
			continue
		}
		runnable = append(runnable, task.callback)
	}
	tw.slots[tw.currentPos] = remaining
	tw.mu.Unlock()

	for _, cb := range runnable {
		go cb()
	}
}
