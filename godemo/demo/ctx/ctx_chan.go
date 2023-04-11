package main

import (
	"fmt"
	"time"
)

var stop chan bool

func reqTask(name string)  {
	for {
		select {
		case <-stop:
			fmt.Println("stop", name)
			break
		default:
			fmt.Println(name, "send request")
			time.Sleep(1*time.Second)
		}
	}
}

func main()  {
	stop = make(chan bool)
	go reqTask("worker")
	time.Sleep(3*time.Second)
	stop <- true
	fmt.Println("main quit")
}
