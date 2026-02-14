package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

const PageSize = 4096

type PageID uint32

// Page 表示固定大小页。
type Page struct {
	ID       PageID
	Data     [PageSize]byte
	Dirty    bool
	PinCount int
}

// PageManager 页管理接口。
type PageManager interface {
	ReadPage(ctx context.Context, id PageID) (*Page, error)
	WritePage(ctx context.Context, page *Page) error
	AllocatePage(ctx context.Context) (PageID, error)
	FreePage(ctx context.Context, id PageID) error
}

var (
	ErrPageNotFound = errors.New("page not found")
)

// FilePageManager 基于文件的页管理器。
type FilePageManager struct {
	mu         sync.Mutex
	file       *os.File
	nextPageID PageID
	freeList   []PageID
}

// NewFilePageManager 创建文件页管理器。
func NewFilePageManager(path string) (*FilePageManager, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	next := PageID(stat.Size() / PageSize)
	return &FilePageManager{file: file, nextPageID: next}, nil
}

// Close 关闭底层文件。
func (m *FilePageManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.file.Close()
}

// ReadPage 读取指定页。
func (m *FilePageManager) ReadPage(ctx context.Context, id PageID) (*Page, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	offset := int64(id) * PageSize
	buf := make([]byte, PageSize)
	n, err := m.file.ReadAt(buf, offset)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if n == 0 && id >= m.nextPageID {
		return nil, ErrPageNotFound
	}
	page := &Page{ID: id}
	copy(page.Data[:], buf)
	return page, nil
}

// WritePage 写入指定页。
func (m *FilePageManager) WritePage(ctx context.Context, page *Page) error {
	if page == nil {
		return errors.New("page is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	offset := int64(page.ID) * PageSize
	if _, err := m.file.WriteAt(page.Data[:], offset); err != nil {
		return err
	}
	if page.ID >= m.nextPageID {
		m.nextPageID = page.ID + 1
	}
	page.Dirty = false
	return nil
}

// AllocatePage 分配新页。
func (m *FilePageManager) AllocatePage(ctx context.Context) (PageID, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.freeList) > 0 {
		id := m.freeList[len(m.freeList)-1]
		m.freeList = m.freeList[:len(m.freeList)-1]
		return id, nil
	}
	id := m.nextPageID
	m.nextPageID++
	return id, nil
}

// FreePage 回收页。
func (m *FilePageManager) FreePage(ctx context.Context, id PageID) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.freeList = append(m.freeList, id)
	return nil
}

func (m *FilePageManager) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprintf("FilePageManager(next=%d, free=%d)", m.nextPageID, len(m.freeList))
}
