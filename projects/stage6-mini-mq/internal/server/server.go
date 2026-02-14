package server

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/protocol"
)

// TCPServer broker TCP 服务。
type TCPServer struct {
	addr    string
	handler *Handler
	logger  *slog.Logger

	listener net.Listener
	wg       sync.WaitGroup
}

// NewTCPServer 创建服务。
func NewTCPServer(addr string, handler *Handler, logger *slog.Logger) *TCPServer {
	return &TCPServer{addr: addr, handler: handler, logger: logger}
}

// Start 启动服务。
func (s *TCPServer) Start(ctx context.Context) error {
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
	s.listener = ln
	s.logger.Info("mini-mq broker started", "addr", s.addr)

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			if ctx.Err() != nil {
				break
			}
			return acceptErr
		}
		s.wg.Add(1)
		go s.handleConn(ctx, conn)
	}
	s.wg.Wait()
	return nil
}

func (s *TCPServer) handleConn(ctx context.Context, conn net.Conn) {
	defer s.wg.Done()
	defer func() { _ = conn.Close() }()

	for {
		if err := conn.SetReadDeadline(time.Now().Add(10 * time.Minute)); err != nil {
			return
		}
		req, err := protocol.DecodeRequest(conn)
		if err != nil {
			return
		}
		resp := s.handler.HandleRequest(ctx, req)
		encoded, err := protocol.EncodeResponse(resp)
		if err != nil {
			return
		}
		if _, err = conn.Write(encoded); err != nil {
			return
		}
	}
}
