package main

import (
	"errors"
	"log"
	"net/http"
	"net/rpc"
)

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (a *Arith) Multiply(args *Args, reply *int) error  {
	*reply = args.A * args.B
	return nil
}

func (a *Arith) Divide(args *Args, quotient *Quotient) error  {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quotient.Quo = args.A / args.B
	quotient.Rem = args.A % args.B
	return nil
}

func main()  {
	serv := rpc.NewServer()
	arith := new(Arith)
	serv.Register(arith)
	err := http.ListenAndServe(":3000", serv)
	if err != nil {
		log.Print(err.Error())
	}
}
