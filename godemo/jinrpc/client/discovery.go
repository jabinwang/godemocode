package client

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const(
	RandomSelect SelectMode = iota

)

type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll()([]string, error)
}

type MultiServerDiscovery struct {
	r *rand.Rand
	mu sync.Mutex
	servers []string
	index int
}

type XClient struct {
	d Discovery
	mode SelectMode
	mu sync.Mutex
	clients map[string]*Client
}

func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	d := &MultiServerDiscovery{
		servers: servers,
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

func (d *MultiServerDiscovery) Refresh() error {
	return nil
}

func (d *MultiServerDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

func (d *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	}
	return "", errors.New("get none")
}

func (d *MultiServerDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	servers := make([]string, len(d.servers), len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}

func NewXClient(d Discovery, mode SelectMode) *XClient {
	return &XClient{
		d: d,
		mode: mode,
		clients: make(map[string]*Client),
	}
}
