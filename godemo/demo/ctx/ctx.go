package main

import (
	"context"
	"fmt"
	"time"
)

type Options struct {
	Interval time.Duration
}

func reqTaskCtx(ctx context.Context, name string)  {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("stop", name)
			return
		default:
			fmt.Println(name, "send request")
			options := ctx.Value("options").(*Options)
			time.Sleep(options.Interval*time.Second)
		}
	}
}
func main()  {
	ctx, cancel := context.WithCancel(context.Background())
	vCtx := context.WithValue(ctx, "options", &Options{2})
	go reqTaskCtx(vCtx, "worker1")
	go reqTaskCtx(vCtx, "worker2")
	time.Sleep(3*time.Second)
	cancel()
	time.Sleep(3*time.Second)
}
