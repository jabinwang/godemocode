package main

import (
	"log"
	"net/rpc"
)
type Args struct{
	A, B int
}

type Quotient struct{
	Quo, Rem int
}

func main()  {
	client, err := rpc.DialHTTP("tcp", ":3000")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	dones := make([]chan *rpc.Call, 0, 10)
	for i := 0; i < 10; i++ {
		quotient := new(Quotient)
		args := &Args{i+10, i}
		divCall := client.Go("Arith.Divide", args, quotient, nil)
		dones = append(dones, divCall.Done)
		log.Print("send", i)

	}
	log.Print("----------------")
	for idx, done := range dones{
		replyCall := <- done
		args := replyCall.Args.(*Args)
		reply := replyCall.Reply.(*Quotient)
		if replyCall.Error != nil {
			log.Println("error ", replyCall.Error)
		}
		log.Printf("%d / %d = %d, %d %% %d = %d\n", args.A, args.B, reply.Quo, args.A, args.B, reply.Rem)
		log.Print("recv", idx)
	}
}
