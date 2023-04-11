package jinrpc

import (
	"encoding/json"
	"errors"
	"io"
	"jinrpc/codec"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

type Option struct {
	CodecType   codec.Type
}

var DefaultOption = &Option{
	CodecType:   codec.GobType,
}

type Server struct {
	serviceMap  sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Accept(lis net.Listener)  {
	for  {
		conn, err := lis.Accept()
		if err != nil {
			return
		}
		go s.ServeConn(conn)
	}
}

func (s Server) ServeConn(conn io.ReadWriteCloser)  {
	defer func() {
		_ = conn.Close()
	}()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		return
	}

	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		return
	}
	s.serveCodec(f(conn))
}

func (s *Server) serveCodec(codec codec.Codec) {
	//每个请求发送完成
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	for  {
		req, err := s.readRequest(codec)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(codec, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(codec, req, sending, wg)
	}
	wg.Wait()
	codec.Close()
}

type request struct {
	h *codec.Header
	argv, replyv reflect.Value
	mtype *methodType
	sev *service
}

var invalidRequest = struct {

}{}

func (s *Server) readRequestHeader(cc codec.Codec) ( *codec.Header,  error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server:  read header error" )
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(code codec.Codec)(*request, error)  {
	h , err := s.readRequestHeader(code)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.sev, req.mtype, err  = s.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err = code.ReadBody(argvi); err != nil {
		log.Println("rpc server read argv error: ", err)
		return req, err
	}
	return req, nil

}

func (s *Server) sendResponse(c codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex )  {
	sending.Lock()
	defer sending.Unlock()
	log.Printf("rpc server response header %v, body %v\n", h, body)
	if err := c.Write(h, body); err != nil {
		log.Println("rpc server: write response error: ", err)
	}

}

func (s *Server) handleRequest(c codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup)  {
	defer wg.Done()
	req.sev.call(req.mtype, req.argv, req.replyv)
	//log.Printf("rpc server %v, %v\n", req.h, req.argv.Elem())
	//req.replyv = reflect.ValueOf(fmt.Sprintf("rpc resp %d", req.h.Seq))
	s.sendResponse(c, req.h, req.replyv.Interface(), sending)
}

func (s *Server) Register(recv interface{}) error {
	sev:= NewService(recv)
	if _, loaded := s.serviceMap.LoadOrStore(sev.name, sev); loaded {
		return errors.New("rpc server: service already defined " + sev.name)
	}
	return nil
}

func (s *Server) findService(serviceMethod string) (sev *service, mtype * methodType, err error)  {
	dot := strings.LastIndex(serviceMethod, ".")
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot +1:]
	sevi, ok := s.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: cannot find service " + serviceName)
	}
	sev = sevi.(*service)
	mtype = sev.method[methodName]
	if mtype == nil {
		err = errors.New("rpc servcer cannot find method " + methodName)
	}
	return
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func Register(recv interface{}) error {
	return DefaultServer.Register(recv)
}
