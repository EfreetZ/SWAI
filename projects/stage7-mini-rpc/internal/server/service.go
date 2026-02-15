package server

import (
	"errors"
	"reflect"
)

var ErrMethodNotFound = errors.New("method not found")

// MethodType 服务方法元信息。
type MethodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

// Service 服务定义。
type Service struct {
	Name    string
	rcvr    reflect.Value
	typ     reflect.Type
	methods map[string]*MethodType
}

// NewService 构建服务并反射注册方法。
func NewService(rcvr interface{}) (*Service, error) {
	s := &Service{rcvr: reflect.ValueOf(rcvr), typ: reflect.TypeOf(rcvr), methods: make(map[string]*MethodType)}
	s.Name = reflect.Indirect(s.rcvr).Type().Name()
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType := mType.In(1)
		replyType := mType.In(2)
		if argType.Kind() != reflect.Ptr || replyType.Kind() != reflect.Ptr {
			continue
		}
		s.methods[method.Name] = &MethodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return s, nil
}

// Call 反射调用服务方法。
func (s *Service) Call(methodName string, args, reply interface{}) error {
	m, ok := s.methods[methodName]
	if !ok {
		return ErrMethodNotFound
	}
	in := []reflect.Value{s.rcvr, reflect.ValueOf(args), reflect.ValueOf(reply)}
	res := m.method.Func.Call(in)
	if errInter := res[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
