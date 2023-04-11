package main

import (
	"net/rpc"
	"sync"
)

const saveQueueLen = 1000

type Store interface {
	Put(url, key *string) error
	Get(key, url *string) error
}

type ProxyStore struct {
	urls *URLStore
	client *rpc.Client
}

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	save chan record
}

type record struct {
	Key, URL string
}

func NewURLStore(filename string) *URLStore  {
	s := &URLStore{urls: make(map[string]string),
		save: make(chan record, saveQueueLen)}
	if filename != "" {
		s.save = make()
	}
	return s
}