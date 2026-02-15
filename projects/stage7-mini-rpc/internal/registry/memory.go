package registry

import "sync"

// MemoryRegistry 内存注册中心。
type MemoryRegistry struct {
	mu       sync.RWMutex
	services map[string][]*ServiceInstance
	watchers map[string][]chan []*ServiceInstance
}

// NewMemoryRegistry 创建内存注册中心。
func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{services: make(map[string][]*ServiceInstance), watchers: make(map[string][]chan []*ServiceInstance)}
}

func (r *MemoryRegistry) Register(instance *ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	instances := r.services[instance.Name]
	for _, v := range instances {
		if v.ID == instance.ID {
			return nil
		}
	}
	r.services[instance.Name] = append(r.services[instance.Name], instance)
	r.notifyWatchers(instance.Name)
	return nil
}

func (r *MemoryRegistry) Deregister(instance *ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	instances := r.services[instance.Name]
	newInstances := make([]*ServiceInstance, 0, len(instances))
	for _, v := range instances {
		if v.ID != instance.ID {
			newInstances = append(newInstances, v)
		}
	}
	r.services[instance.Name] = newInstances
	r.notifyWatchers(instance.Name)
	return nil
}

func (r *MemoryRegistry) Discover(serviceName string) ([]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	instances := r.services[serviceName]
	res := make([]*ServiceInstance, len(instances))
	copy(res, instances)
	return res, nil
}

func (r *MemoryRegistry) Watch(serviceName string) (<-chan []*ServiceInstance, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ch := make(chan []*ServiceInstance, 1)
	r.watchers[serviceName] = append(r.watchers[serviceName], ch)
	ch <- append([]*ServiceInstance(nil), r.services[serviceName]...)
	return ch, nil
}

func (r *MemoryRegistry) notifyWatchers(serviceName string) {
	watchers := r.watchers[serviceName]
	snapshot := append([]*ServiceInstance(nil), r.services[serviceName]...)
	for _, ch := range watchers {
		select {
		case ch <- snapshot:
		default:
		}
	}
}
