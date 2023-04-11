package main

import (
	"flag"
	"fmt"
	"jincache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *jincache.Group {
	return jincache.NewGroup("scores", 2<<10, jincache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[db] search key in local", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, group *jincache.Group) {
	httpPool := jincache.NewHttpPool(addr)
	httpPool.Set(addrs...)
	group.RegisterPeers(httpPool)
	log.Println("jincache is running at ", addr)
	//http://localhost:9999/jin_cache/scores/Tom
	log.Fatal(http.ListenAndServe(addr[7:], httpPool))
}

// http://localhost:9999/api?key=Tom
func startApiServer(apiAddr string, group *jincache.Group) {
	http.Handle("/api", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")
		view, err := group.Get(key)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("content-type", "application/octet-stream")
		writer.Write(view.ByteSlice())
	}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "jincache server port")
	flag.BoolVar(&api, "api", false, "start a api server")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	group := createGroup()
	if api {
		go startApiServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], addrs, group)
}
