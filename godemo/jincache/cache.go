package jincache

import (
	"jincache/lru"
	"sync"
)

type safeCache struct {
	mu sync.Mutex
	lru *lru.Cache
	cacheBytes int64
}

func (c *safeCache) add(key string, value ByteView)  {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *safeCache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
