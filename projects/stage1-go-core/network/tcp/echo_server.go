package tcp

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
)

// EchoServer 提供最小可用的 TCP Echo 服务。
type EchoServer struct {
	addr string

	mu       sync.RWMutex
	listener net.Listener
	wg       sync.WaitGroup
}

// NewEchoServer 创建一个新的 EchoServer。
func NewEchoServer(addr string) *EchoServer {
	return &EchoServer{addr: addr}
}

// Addr 返回当前监听地址。
func (s *EchoServer) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

// Start 启动服务，并在 ctx 取消时优雅退出。
func (s *EchoServer) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()

	defer func() {
		_ = ln.Close()
		s.wg.Wait()
	}()

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			if ctx.Err() != nil || errors.Is(acceptErr, net.ErrClosed) {
				return nil
			}
			return acceptErr
		}

		s.wg.Add(1)
		go s.handleConn(ctx, conn)
	}
}

func (s *EchoServer) handleConn(ctx context.Context, conn net.Conn) {
	defer s.wg.Done()
	defer func() {
		_ = conn.Close()
	}()

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()

	_, _ = io.Copy(conn, conn)
	close(done)
}
