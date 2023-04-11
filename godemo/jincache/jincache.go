package jincache

import (
	"fmt"
	"jincache/pb/pb"
	"jincache/singleflight"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string)([]byte, error)

func (f GetterFunc) Get(key string)([]byte, error)  {
	return f(key)
}

type Group struct {
	 name string
	 getter Getter
	 mainCache safeCache
	 peers PeerPicker
	 loader *singleflight.Single
}

var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group  {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name: name,
		getter: getter,
		mainCache: safeCache{cacheBytes: cacheBytes},
		loader: &singleflight.Single{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group  {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string)(ByteView, error)  {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is nil")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("cache hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	log.Println("cache not hit and load from local or remote")
	byteView , err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peerGetter, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peerGetter, key); err == nil {
					return value, nil
				}
			}
		}
		return g.loadFromLocal(key)
	})

	if err == nil {
		return byteView.(ByteView), nil
	}
	return ByteView{}, nil
}
func (g Group) getFromPeer(peer PeerGetter, key string)(ByteView, error)  {
	req := &pb.Request{
		Group: g.name,
		Key: key,
	}
	res := &pb.Response{

	}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

//func (g *Group) getFromPeer(peerGetter PeerGetter, key string)(ByteView, error)  {
//	log.Println("cache not hit and get from peer")
//	bytes, err := peerGetter.Get(g.name, key)
//	if err != nil {
//		log.Printf("get %s failed in peer", key)
//		return ByteView{}, err
//	}
//	return ByteView{b: bytes}, nil
//}

func (g *Group) loadFromLocal(key string) (ByteView, error)  {
	log.Println("cache not hit and get from local")
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{
		b: cloneBytes(bytes),
	}

	g.mainCache.add(key, value)
	return value, nil
}

func (g *Group) RegisterPeers(peers PeerPicker)  {
	if g.peers != nil {
		panic("RegsiterPeers call more than one")
	}
	g.peers = peers
}


