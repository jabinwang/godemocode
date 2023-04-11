package jinrpc

import (
	"reflect"
	"sync/atomic"
)
/**
func (t *T) MethodName(argType T1, replyType *T2) error
 */

type methodType struct {
	method reflect.Method
	ArgType reflect.Type
	ReplyType reflect.Type
	numCalls uint64
}

func (m *methodType) NumCalls() uint64  {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

type service struct {
	name string
	typ reflect.Type
	recv reflect.Value
	method map[string]*methodType
}

func NewService(recv interface{}) *service  {
	s := new(service)
	s.recv = reflect.ValueOf(recv)
	s.name = reflect.Indirect(s.recv).Type().Name()
	s.typ = reflect.TypeOf(recv)
	s.registerMethods()
	return s
}

func (s *service) registerMethods()  {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		ty := method.Type
		argType, replyType := ty.In(1), ty.In(2)
		s.method[method.Name] = &methodType{
			method: method,
			ArgType: argType,
			ReplyType: replyType,
		}
	}
}

func (s *service ) call(m *methodType, argv, replyv reflect.Value) error  {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnVaues := f.Call([]reflect.Value{s.recv, argv, replyv})
	if err := returnVaues[0].Interface(); err != nil {
		return err.(error)
	}
	return nil
}


