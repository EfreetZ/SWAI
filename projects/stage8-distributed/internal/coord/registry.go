package coord

import "sync"

// ServiceInstance 服务实例。
type ServiceInstance struct {
	ID   string
	Name string
	Addr string
}

// Registry 服务注册发现。
type Registry struct {
	mu       sync.RWMutex
	services map[string][]ServiceInstance
}

// NewRegistry 创建注册中心。
func NewRegistry() *Registry {
	return &Registry{services: make(map[string][]ServiceInstance)}
}

// Register 注册实例。
func (r *Registry) Register(instance ServiceInstance) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[instance.Name] = append(r.services[instance.Name], instance)
}

// Discover 发现实例。
func (r *Registry) Discover(name string) []ServiceInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := r.services[name]
	res := make([]ServiceInstance, len(items))
	copy(res, items)
	return res
}
