package main

import (
	"fmt"
	"jinCron"
	"sync"
	"time"
)

func main()  {
	var wg sync.WaitGroup
	wg.Add(1)
	c := jinCron.NewCron()
	c.AddFunc(jinCron.Every(1*time.Second), func() {
		fmt.Println("Every 1 second")
	})
	c.AddFunc(jinCron.Every(5*time.Second), func() {
		fmt.Println("Every 5 second")
	})
	c.Start()
	wg.Wait()

	//var wg sync.WaitGroup
	//wg.Add(1)
	//c := gron.New()
	//c.AddFunc(gron.Every(1*time.Second), func() {
	//	fmt.Println("runs every hour.")
	//})
	//c.Start()
	//wg.Wait()
}
