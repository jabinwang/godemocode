package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/golang/groupcache"
	"log"
	"net/http"
	"strings"
)
var Store = map[string][]byte{
	"red": []byte("#FF0000"),
	"green": []byte("#00FF00"),
	"blue": []byte("#0000FF"),
}

var Group = groupcache.NewGroup("foobar", 64<<20, groupcache.GetterFunc(
	func(ctx context.Context, key string, dest groupcache.Sink) error {
		log.Println("looking up", key)
		v, ok := Store[key]
		if !ok {
			return errors.New("color not found")
		}
		dest.SetBytes(v)
		return nil
	}))

func main()  {
	addr := flag.String("addr", ":8080", "server address")
	peers := flag.String("pool", "http://localhost:8080", "server pool list")
	flag.Parse()
	http.HandleFunc("/color", func(responseWriter http.ResponseWriter, request *http.Request) {
		color := request.FormValue("name")
		var b []byte
		err := Group.Get(nil, color, groupcache.AllocatingByteSliceSink(&b))
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusNotFound)
		}
		responseWriter.Write(b)
		responseWriter.Write([]byte{'\n'})
	})
	p := strings.Split(*peers, ",")
	fmt.Println("p ", p)
	pool := groupcache.NewHTTPPool(p[0])
	pool.Set(p...)
	http.ListenAndServe(*addr, nil)
}