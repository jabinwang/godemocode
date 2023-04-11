package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Address struct {
	Type string
	City string
	Country string
}

type VCard struct {
	FirstName string
	LastName string
	Addresses []*Address
	Remark string
}

func main()  {
	pa := &Address{"private", "beijing", "cn"}
	wa := &Address{"work", "xiamen", "cn"}

	vc := VCard{"Jan", "wang", []*Address{pa, wa}, "none"}
	js, _ := json.Marshal(vc)
	fmt.Printf("json format: %s", js)
	file, _ := os.OpenFile("vcard.json", os.O_CREATE | os.O_WRONLY, 0666)
	defer file.Close()
	enc := json.NewEncoder(file)
	err := enc.Encode(vc)
	if err != nil {
		log.Println("Error in Encoding json")
	}

}
