package registry

import "errors"

var ErrNotImplemented = errors.New("etcd registry not implemented in stage7")

// EtcdRegistry 占位实现，后续阶段对接 etcd。
type EtcdRegistry struct{}

func (e *EtcdRegistry) Register(instance *ServiceInstance) error {
	return ErrNotImplemented
}

func (e *EtcdRegistry) Deregister(instance *ServiceInstance) error {
	return ErrNotImplemented
}

func (e *EtcdRegistry) Discover(serviceName string) ([]*ServiceInstance, error) {
	return nil, ErrNotImplemented
}

func (e *EtcdRegistry) Watch(serviceName string) (<-chan []*ServiceInstance, error) {
	return nil, ErrNotImplemented
}
