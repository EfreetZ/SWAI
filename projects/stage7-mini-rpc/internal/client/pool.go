package client

import (
	"errors"
	"net"
	"sync"
)

// ConnPool 简单连接池。
type ConnPool struct {
	factory   func() (net.Conn, error)
	conns     chan net.Conn
	maxIdle   int
	maxActive int
	active    int
	mu        sync.Mutex
	closed    bool
}

// NewConnPool 创建连接池。
func NewConnPool(maxIdle, maxActive int, factory func() (net.Conn, error)) *ConnPool {
	if maxIdle <= 0 {
		maxIdle = 2
	}
	if maxActive <= 0 {
		maxActive = 10
	}
	if maxIdle > maxActive {
		maxIdle = maxActive
	}
	return &ConnPool{factory: factory, conns: make(chan net.Conn, maxIdle), maxIdle: maxIdle, maxActive: maxActive}
}

// Get 获取连接。
func (p *ConnPool) Get() (net.Conn, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("pool closed")
	}
	select {
	case conn := <-p.conns:
		p.mu.Unlock()
		return conn, nil
	default:
	}
	if p.active >= p.maxActive {
		p.mu.Unlock()
		conn := <-p.conns
		return conn, nil
	}
	p.active++
	p.mu.Unlock()

	conn, err := p.factory()
	if err != nil {
		p.mu.Lock()
		p.active--
		p.mu.Unlock()
		return nil, err
	}
	return conn, nil
}

// Put 归还连接。
func (p *ConnPool) Put(conn net.Conn) error {
	if conn == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		_ = conn.Close()
		p.active--
		return nil
	}
	select {
	case p.conns <- conn:
		return nil
	default:
		_ = conn.Close()
		p.active--
		return nil
	}
}

// Close 关闭连接池。
func (p *ConnPool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.conns)
	for conn := range p.conns {
		_ = conn.Close()
	}
	p.mu.Unlock()
	return nil
}
