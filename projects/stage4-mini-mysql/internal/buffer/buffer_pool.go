package buffer

import (
	"context"
	"errors"
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
)

var ErrNoFrameAvailable = errors.New("no frame available")

// BufferPoolManager Buffer Pool 管理器。
type BufferPoolManager struct {
	mu       sync.Mutex
	capacity int
	disk     storage.PageManager
	replacer Replacer

	pages      map[storage.PageID]*storage.Page
	frameByPID map[storage.PageID]FrameID
	pidByFrame map[FrameID]storage.PageID
	freeList   []FrameID
}

// NewBufferPoolManager 创建 Buffer Pool。
func NewBufferPoolManager(capacity int, disk storage.PageManager) *BufferPoolManager {
	if capacity <= 0 {
		capacity = 64
	}
	freeList := make([]FrameID, 0, capacity)
	for i := 0; i < capacity; i++ {
		freeList = append(freeList, FrameID(i))
	}
	return &BufferPoolManager{
		capacity:   capacity,
		disk:       disk,
		replacer:   NewLRUReplacer(),
		pages:      make(map[storage.PageID]*storage.Page),
		frameByPID: make(map[storage.PageID]FrameID),
		pidByFrame: make(map[FrameID]storage.PageID),
		freeList:   freeList,
	}
}

// FetchPage 获取页。
func (b *BufferPoolManager) FetchPage(ctx context.Context, id storage.PageID) (*storage.Page, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	b.mu.Lock()
	if p, ok := b.pages[id]; ok {
		p.PinCount++
		b.replacer.Pin(b.frameByPID[id])
		b.mu.Unlock()
		return p, nil
	}
	frameID, ok := b.acquireFrameLocked(ctx)
	if !ok {
		b.mu.Unlock()
		return nil, ErrNoFrameAvailable
	}
	b.mu.Unlock()

	page, err := b.disk.ReadPage(ctx, id)
	if err != nil {
		return nil, err
	}

	b.mu.Lock()
	page.PinCount = 1
	b.pages[id] = page
	b.frameByPID[id] = frameID
	b.pidByFrame[frameID] = id
	b.replacer.Pin(frameID)
	b.mu.Unlock()
	return page, nil
}

// UnpinPage 取消 pin。
func (b *BufferPoolManager) UnpinPage(ctx context.Context, id storage.PageID, isDirty bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	page, ok := b.pages[id]
	if !ok {
		return storage.ErrPageNotFound
	}
	if page.PinCount > 0 {
		page.PinCount--
	}
	if isDirty {
		page.Dirty = true
	}
	if page.PinCount == 0 {
		b.replacer.Unpin(b.frameByPID[id])
	}
	return nil
}

// FlushPage 刷盘单页。
func (b *BufferPoolManager) FlushPage(ctx context.Context, id storage.PageID) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	page, ok := b.pages[id]
	b.mu.Unlock()
	if !ok {
		return storage.ErrPageNotFound
	}
	return b.disk.WritePage(ctx, page)
}

// NewPage 新建页。
func (b *BufferPoolManager) NewPage(ctx context.Context) (*storage.Page, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	pageID, err := b.disk.AllocatePage(ctx)
	if err != nil {
		return nil, err
	}
	page := &storage.Page{ID: pageID, Dirty: true, PinCount: 1}

	b.mu.Lock()
	frameID, ok := b.acquireFrameLocked(ctx)
	if !ok {
		b.mu.Unlock()
		return nil, ErrNoFrameAvailable
	}
	b.pages[pageID] = page
	b.frameByPID[pageID] = frameID
	b.pidByFrame[frameID] = pageID
	b.replacer.Pin(frameID)
	b.mu.Unlock()
	return page, nil
}

// FlushAllPages 刷新全部页。
func (b *BufferPoolManager) FlushAllPages(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	ids := make([]storage.PageID, 0, len(b.pages))
	for id := range b.pages {
		ids = append(ids, id)
	}
	b.mu.Unlock()

	for _, id := range ids {
		if err := b.FlushPage(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (b *BufferPoolManager) acquireFrameLocked(ctx context.Context) (FrameID, bool) {
	if len(b.freeList) > 0 {
		last := len(b.freeList) - 1
		frameID := b.freeList[last]
		b.freeList = b.freeList[:last]
		return frameID, true
	}
	victim, ok := b.replacer.Victim()
	if !ok {
		return 0, false
	}
	pid, exists := b.pidByFrame[victim]
	if !exists {
		return victim, true
	}
	page := b.pages[pid]
	if page != nil && page.Dirty {
		_ = b.disk.WritePage(ctx, page)
	}
	delete(b.pages, pid)
	delete(b.frameByPID, pid)
	delete(b.pidByFrame, victim)
	return victim, true
}
