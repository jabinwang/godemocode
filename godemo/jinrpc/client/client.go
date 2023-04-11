package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"jinrpc"
	"jinrpc/codec"
	"log"
	"net"
	"sync"
)

type Call struct {
	Seq uint64
	ServiceMethod string
	Args interface{}
	Reply interface{}
	Error error
	Done chan *Call
}

func (c *Call) done()  {
	c.Done <- c
}

type Client struct {
	cc codec.Codec
	opt *jinrpc.Option
	sending *sync.Mutex
	header codec.Header
	mu sync.Mutex
	seq uint64
	pending map[uint64]*Call
	closing bool
	shutdown bool
}
var ErrShutdown = errors.New("conn is shut down")
func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

func  NewClient(conn net.Conn) (*Client, error) {
	opt := jinrpc.DefaultOption
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		conn.Close()
		return nil, err
	}
	client := &Client{
		seq: 1,
		cc: f(conn),
		opt: opt,
		pending: make(map[uint64]*Call),
		sending: new(sync.Mutex),
	}
	go client.receive()
	return client, nil
}

func (c *Client) receive()  {
	var err error
	for err == nil {
		var h codec.Header
		if err := c.cc.ReadHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch  {
		case call == nil:
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body error")
			}
			call.done()
		}
	}

}

func Dial(network, addr string)(client *Client, err error)  {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if client == nil {
			conn.Close()
		}
	}()

	return NewClient(conn)
}

func (c *Client) send(call *Call)  {
	c.sending.Lock()
	defer c.sending.Unlock()

	seq, err := c.registerCall(call)

	if err != nil {
		call.Error = err
		call.done()
		return
	}

	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	c.header.Error = ""

	log.Printf("rpc client: req header %v, call seq %v,  args %v\n", c.header, call.Seq, call.Args)
	if err := c.cc.Write(&c.header, call.Args); err != nil {
		call := c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (c *Client) Go(serviceMethod string, args , reply interface{}, done chan *Call) *Call  {
	if done == nil {
		done = make(chan *Call, 10)
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args: args,
		Reply: reply,
		Done: done,
	}
	c.send(call)
	return call
}

func (c *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <- c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}
