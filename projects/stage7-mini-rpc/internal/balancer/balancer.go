package balancer

import (
	"errors"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
)

var ErrNoInstance = errors.New("no service instance")

// Balancer 负载均衡接口。
type Balancer interface {
	Pick(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error)
}
