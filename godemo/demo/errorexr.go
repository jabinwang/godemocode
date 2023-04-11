package main

import (
	"errors"
	"fmt"
)

func hello(name string) error {
	if len(name) == 0 {
		return errors.New("error: name is null")
	}
	fmt.Println("hello", name)
	return nil
}

func  get(index int) (ret int)  {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("some error happened", r)
			ret = -1
		}
	}()

	arr := [3]int{1,2,3}
	return arr[index]
}

func main()  {
	//if err := hello("test"); err != nil {
	//	fmt.Println(err)
	//}
	fmt.Println(get(5))
	fmt.Println("finished")
}
