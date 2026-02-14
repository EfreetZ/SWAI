package server

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/executor"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/parser"
)

// TCPServer TCP 文本协议服务。
type TCPServer struct {
	addr   string
	engine *executor.Engine
	logger *slog.Logger

	listener net.Listener
	wg       sync.WaitGroup
	nextID   atomic.Uint64
}

// NewTCPServer 创建 TCP 服务。
func NewTCPServer(addr string, engine *executor.Engine, logger *slog.Logger) *TCPServer {
	return &TCPServer{addr: addr, engine: engine, logger: logger}
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
	s.logger.Info("mini-mysql tcp server started", "addr", s.addr)

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
	defer func() {
		_ = conn.Close()
	}()

	sessionID := fmt.Sprintf("session-%d", s.nextID.Add(1))
	_, _ = conn.Write([]byte("WELCOME mini-mysql\n"))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := normalizeSQL(scanner.Text())
		if line == "" {
			continue
		}
		if line == "QUIT" || line == "EXIT" {
			_, _ = conn.Write([]byte("BYE\n"))
			return
		}

		statement, err := parser.Parse(line)
		if err != nil {
			_, _ = conn.Write([]byte("ERR invalid sql\n"))
			continue
		}

		opCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		result, execErr := s.engine.Execute(opCtx, sessionID, statement)
		cancel()
		if execErr != nil {
			_, _ = conn.Write([]byte("ERR " + execErr.Error() + "\n"))
			continue
		}
		_, _ = conn.Write([]byte("OK " + result + "\n"))
	}
}
