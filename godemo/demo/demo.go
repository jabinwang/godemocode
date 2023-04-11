package main

import (
	"log"
	"time"
)

type Books struct {
	title string
	author string
	subject string
	book_id int
}

func StartTimer(name string) func() {
	t := time.Now()
	log.Println(name, "started")
	return func() {
		d := time.Now().Sub(t)
		log.Println(name, "took", d)
	}
}

func RunTimer()  {
	stop := StartTimer("My Timer")
	defer stop()
	time.Sleep(1 * time.Second)
}

func main() {
	//fmt.Println(Books{"go", "lee","go",111})
	RunTimer()
}
