package lru

import "container/list"

type Cache struct {
	maxBytes int64
	nbytes int64
	OnEvicted func(key string, value Value)
	ll *list.List
	cache map[string]*list.Element
}

type entry struct {
	key string
	value Value
}

type Value interface {
	Len() int
}

func New(max int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes: max,
		ll: list.New(),
		cache: make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Add(key string, value Value)  {
	if c.cache == nil {
		c.cache = make(map[string]*list.Element)
		c.ll = list.New()
	}
	if e, ok := c.cache[key]; ok {
		c.ll.MoveToFront(e)
		kv := e.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	c.nbytes += int64(len(key)) + int64(value.Len())
	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.removeOldest()
	}
}

func (c *Cache) removeOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(ele *list.Element) {
	c.ll.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.cache, kv.key)
	c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}


func (c *Cache) Get(key string) (value Value, ok bool) {
	if c.cache == nil {
		return
	}
	if ele , hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}



