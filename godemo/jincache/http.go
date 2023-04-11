package jincache

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"jincache/pb/pb"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_jincache/"
	defaultReplicas = 50
)

// HttpPool /basepath/groupname/key
type HttpPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *Map
	httpGetters map[string]*httpGetter
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}



func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HttpPool serving invalid path" + r.URL.Path)
	}
	log.Printf("server %s %s\n", r.Method, r.URL.Path)
	// /basepath/groupname/key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group"+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")

	w.Write(body)
	//w.Write(view.ByteSlice())
}

func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = NewMap(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseUrl: peer + p.basePath}
	}
}

func (p *HttpPool) PickPeer(key string) ( PeerGetter,  bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		log.Printf("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

type httpGetter struct {
	baseUrl string
}

//func (h *httpGetter) Get(group string, key string) ([]byte, error) {
//	u := fmt.Sprintf("%v%v/%v",
//		h.baseUrl,
//		url.QueryEscape(group),
//		url.QueryEscape(key))
//	log.Printf("peer url : %v\n", u)
//	res, err := http.Get(u)
//	if err != nil {
//		return nil, err
//	}
//
//	defer res.Body.Close()
//	if res.StatusCode != http.StatusOK {
//		return nil, fmt.Errorf("server return %v", res.StatusCode)
//	}
//	bytes, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		return nil, fmt.Errorf("reading response body error")
//	}
//
//	return bytes, nil
//}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf("%v%v/%v",
		h.baseUrl,
		url.QueryEscape(in.Group),
		url.QueryEscape(in.Key))
	log.Printf("peer url : %v\n", u)
	res, err := http.Get(u)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server return %v", res.StatusCode)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body error")
	}
	if err := proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding res body error %v", err)
	}

	return nil
}
