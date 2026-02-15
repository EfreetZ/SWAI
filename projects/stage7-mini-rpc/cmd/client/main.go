package main

import (
	"context"
	"fmt"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/client"
)

type AddArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddReply struct {
	Sum int `json:"sum"`
}

func main() {
	c, err := client.Dial("127.0.0.1:18080")
	if err != nil {
		return
	}
	defer func() { _ = c.Close() }()

	proxy := client.NewProxy(c, "ArithService")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var reply AddReply
	_ = proxy.Invoke(ctx, "Add", &AddArgs{A: 1, B: 2}, &reply)
	fmt.Println(reply.Sum)
}
