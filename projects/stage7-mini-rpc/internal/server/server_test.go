package server

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
)

type addArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

type addReply struct {
	Sum int `json:"sum"`
}

type testArith struct{}

func (s *testArith) Add(args *addArgs, reply *addReply) error {
	if args == nil || reply == nil {
		return errors.New("nil args")
	}
	reply.Sum = args.A + args.B
	return nil
}

func TestRegisterService(t *testing.T) {
	srv := NewServer("127.0.0.1:0", registry.NewMemoryRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := srv.Register(&testArith{}); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if _, ok := srv.services["testArith"]; !ok {
		t.Fatal("service not registered")
	}
}
