package main

import (
	"errors"
	"io"
	"log/slog"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/server"
)

type AddArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddReply struct {
	Sum int `json:"sum"`
}

type ArithService struct{}

func (s *ArithService) Add(args *AddArgs, reply *AddReply) error {
	if args == nil || reply == nil {
		return errors.New("invalid args or reply")
	}
	reply.Sum = args.A + args.B
	return nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(io.Writer(io.Discard), nil))
	reg := registry.NewMemoryRegistry()
	srv := server.NewServer("127.0.0.1:18080", reg, logger)
	srv.Use(server.LoggingMiddleware(logger), server.RecoveryMiddleware(logger), server.RateLimitMiddleware(5000))
	_ = srv.Register(&ArithService{})
	_ = srv.Start()
}
