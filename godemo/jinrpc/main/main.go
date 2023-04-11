package main

import (
	"jinrpc"
	"jinrpc/client"
	"log"
	"net"
	"sync"
	"time"
)

func startServer(addr chan string)  {
	var foo Foo
	jinrpc.Register(&foo)

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error: ", err)
	}
	log.Println("start rpc server on ", l.Addr())
	addr <- l.Addr().String()
	jinrpc.Accept(l)
}

func newClient(addr string)  {
	cli, _ := client.Dial("tcp", addr)
	defer func() {

	}()
	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i:= 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i*i}
			var reply int
			if err := cli.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
	}
	wg.Wait()
}

type Foo int
type Args struct {
	Num1, Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func main()  {
	addr := make(chan string)
	go startServer(addr)

	//conn, _ := net.Dial("tcp", <-addr)
	//defer func() {
	//	conn.Close()
	//}()
	//// send options
	//_ = json.NewEncoder(conn).Encode(jinrpc.DefaultOption)
	//cc := codec.NewGobCodec(conn)
	//// send request & receive response
	//for i := 0; i < 5; i++ {
	//	h := &codec.Header{
	//		ServiceMethod: "Foo.Sum",
	//		Seq:           uint64(i),
	//	}
	//	_ = cc.Write(h, fmt.Sprintf("jinrpc req %d", h.Seq))
	//	_ = cc.ReadHeader(h)
	//	var reply string
	//	_ = cc.ReadBody(&reply)
	//	log.Println("reply:", reply)
	//}

	//cli, _ := client.Dial("tcp", <-addr)
	//defer func() {
	//
	//}()
	//time.Sleep(time.Second)
	//var wg sync.WaitGroup
	//for i:= 0; i < 5; i++ {
	//	wg.Add(1)
	//	go func(i int) {
	//		defer wg.Done()
	//		args := fmt.Sprintf("jinrpc req %d", i)
	//		var reply string
	//		if err := cli.Call("Foo.Sum", args, &reply); err != nil {
	//			log.Fatal("call Foo.Sum error:", err)
	//		}
	//		log.Println("reply:", reply)
	//	}(i)
	//}
	//wg.Wait()
	newClient(<-addr)
}
