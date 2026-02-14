package server

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/protocol"
	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/ttl"
)

// TCPServer mini-redis TCP 服务。
type TCPServer struct {
	addr       string
	database   *db.DB
	ttlManager *ttl.Manager
	logger     *slog.Logger

	listener net.Listener
	wg       sync.WaitGroup
}

// NewTCPServer 创建 TCP 服务。
func NewTCPServer(addr string, database *db.DB, ttlManager *ttl.Manager, logger *slog.Logger) *TCPServer {
	return &TCPServer{addr: addr, database: database, ttlManager: ttlManager, logger: logger}
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
	s.logger.Info("mini-redis server started", "addr", s.addr)

	go s.ttlManager.Start(ctx, 100*time.Millisecond)
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

	reader := bufio.NewReader(conn)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(10 * time.Minute)); err != nil {
			return
		}
		value, err := protocol.Parse(reader)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "closed") {
				return
			}
			_, _ = conn.Write(protocol.Serialize(&protocol.Value{Type: protocol.ErrorType, Str: "ERR invalid protocol"}))
			continue
		}
		if value.Type != protocol.Array {
			_, _ = conn.Write(protocol.Serialize(&protocol.Value{Type: protocol.ErrorType, Str: "ERR command must be array"}))
			continue
		}

		args := make([]string, 0, len(value.Array))
		for _, item := range value.Array {
			args = append(args, item.Str)
		}

		result, execErr := s.database.ExecuteCommand(ctx, args)
		if execErr != nil {
			_, _ = conn.Write(protocol.Serialize(&protocol.Value{Type: protocol.ErrorType, Str: "ERR " + execErr.Error()}))
			continue
		}
		_, _ = conn.Write(protocol.Serialize(&protocol.Value{Type: protocol.BulkString, Str: result}))
	}
}
