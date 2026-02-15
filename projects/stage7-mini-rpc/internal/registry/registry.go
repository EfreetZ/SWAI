package registry

// ServiceInstance 服务实例。
type ServiceInstance struct {
	ID       string
	Name     string
	Addr     string
	Metadata map[string]string
	Weight   int
}

// Registry 注册中心接口。
type Registry interface {
	Register(instance *ServiceInstance) error
	Deregister(instance *ServiceInstance) error
	Discover(serviceName string) ([]*ServiceInstance, error)
	Watch(serviceName string) (<-chan []*ServiceInstance, error)
}
