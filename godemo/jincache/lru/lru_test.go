package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestEvict(t *testing.T)  {
	keys := make([]string, 0)
	onEvictFun := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), onEvictFun)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))
	lru.Add("k5", String("k5"))

	expect := []string{"key1", "k2", "k3"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("failed")
	}
}
