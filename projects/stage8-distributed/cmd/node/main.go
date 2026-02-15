package main

import (
	"context"
	"fmt"

	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/kv"
	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/raft"
)

func main() {
	cluster := raft.NewCluster([]string{"n1", "n2", "n3"})
	svc := kv.NewService(cluster)
	_ = svc.Put(context.Background(), "hello", "world")
	v, _ := svc.Get("hello")
	fmt.Println(v)
}
