package main

import (
	"fmt"
	"time"
)

var ch = make(chan string, 10)

func download(url string)  {
	fmt.Println("start to download", url)
	time.Sleep(time.Second)
	ch <- url
}

func main()  {
	for i := 0; i < 3; i++ {
		go download("a.com/" + string(rune(i+'0')))
	}
	for i := 0; i < 3; i++ {
		msg := <- ch
		fmt.Println("finished", msg)
	}
	fmt.Println("Done")
}
