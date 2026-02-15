package server

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/protocol"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
	ErrInternal       = errors.New("internal server error")
	ErrRateLimited    = errors.New("rate limited")
)

// Server RPC 服务端。
type Server struct {
	addr        string
	listener    net.Listener
	services    map[string]*Service
	middlewares []Middleware
	registry    registry.Registry
	instance    *registry.ServiceInstance
	logger      *slog.Logger
	conns       map[net.Conn]struct{}
	mu          sync.RWMutex
}

// NewServer 创建服务端。
func NewServer(addr string, reg registry.Registry, logger *slog.Logger) *Server {
	return &Server{addr: addr, services: make(map[string]*Service), registry: reg, logger: logger, conns: make(map[net.Conn]struct{})}
}

// Register 注册服务。
func (s *Server) Register(rcvr interface{}) error {
	service, err := NewService(rcvr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.services[service.Name] = service
	s.mu.Unlock()
	return nil
}

// Use 添加中间件。
func (s *Server) Use(mw ...Middleware) {
	s.middlewares = append(s.middlewares, mw...)
}

// Start 启动服务。
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	if s.registry != nil {
		s.mu.Lock()
		s.instance = &registry.ServiceInstance{ID: s.addr, Name: "rpc-server", Addr: s.addr, Weight: 1}
		instance := s.instance
		s.mu.Unlock()
		_ = s.registry.Register(instance)
	}

	for {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			if errors.Is(acceptErr, net.ErrClosed) {
				return nil
			}
			return acceptErr
		}
		s.mu.Lock()
		s.conns[conn] = struct{}{}
		s.mu.Unlock()
		go s.handleConnection(conn)
	}
}

// Stop 停止服务。
func (s *Server) Stop() error {
	s.mu.Lock()
	instance := s.instance
	s.mu.Unlock()

	if s.registry != nil && instance != nil {
		_ = s.registry.Deregister(instance)
	}
	s.mu.Lock()
	for conn := range s.conns {
		_ = conn.Close()
		delete(s.conns, conn)
	}
	s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.conns, conn)
		s.mu.Unlock()
		_ = conn.Close()
	}()
	reader := bufio.NewReader(conn)
	for {
		headerBytes := make([]byte, protocol.HeaderSize)
		if _, err := io.ReadFull(reader, headerBytes); err != nil {
			return
		}
		header := protocol.DecodeHeader(headerBytes)
		if header.Magic != protocol.MagicNumber {
			return
		}
		payload := make([]byte, header.PayloadLength)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return
		}

		codec := protocol.GetCodec(header.Codec)
		req := &protocol.RPCRequest{}
		if err := codec.Decode(payload, req); err != nil {
			_ = s.writeResp(conn, header.RequestID, codec, err.Error(), nil)
			continue
		}
		ctx := &Context{Service: req.ServiceName, Method: req.MethodName, Metadata: req.Metadata, StartTime: time.Now()}
		finalHandler := func(c *Context) error {
			data, err := s.invoke(req)
			if err != nil {
				return err
			}
			c.Response = data
			return nil
		}
		chain := ChainMiddleware(finalHandler, s.middlewares...)
		err := chain(ctx)
		if err != nil {
			_ = s.writeResp(conn, header.RequestID, codec, err.Error(), nil)
			continue
		}
		_ = s.writeResp(conn, header.RequestID, codec, "", ctx.Response)
	}
}

func (s *Server) invoke(req *protocol.RPCRequest) ([]byte, error) {
	s.mu.RLock()
	service, ok := s.services[req.ServiceName]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrMethodNotFound
	}
	method, ok := service.methods[req.MethodName]
	if !ok {
		return nil, ErrMethodNotFound
	}

	codec := protocol.GetCodec(protocol.JSON)
	args := reflect.New(method.ArgType.Elem()).Interface()
	reply := reflect.New(method.ReplyType.Elem()).Interface()
	if err := codec.Decode(req.Args, args); err != nil {
		return nil, err
	}
	if err := service.Call(req.MethodName, args, reply); err != nil {
		return nil, err
	}
	return codec.Encode(reply)
}

func (s *Server) writeResp(conn net.Conn, reqID uint64, codec protocol.Codec, errMsg string, data []byte) error {
	resp := &protocol.RPCResponse{RequestID: reqID, Error: errMsg, Data: data}
	payload, err := codec.Encode(resp)
	if err != nil {
		return err
	}
	h := protocol.Header{Magic: protocol.MagicNumber, Version: 1, Type: protocol.Response, Codec: protocol.JSON, RequestID: reqID, PayloadLength: uint32(len(payload))}
	if _, err = conn.Write(protocol.EncodeHeader(h)); err != nil {
		return err
	}
	_, err = conn.Write(payload)
	return err
}
