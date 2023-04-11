package jincache

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int
	hashKeys     []int
	hashMap   map[int]string
}

func NewMap(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:   make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.hashKeys = append(m.hashKeys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.hashKeys)
}

func (m *Map) Get(key string) string {
	if len(m.hashKeys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.hashKeys), func(i int) bool {
		return m.hashKeys[i] >= hash
	})

	return m.hashMap[m.hashKeys[idx%len(m.hashKeys)]]
}
