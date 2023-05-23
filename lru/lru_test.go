package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestAdd(t *testing.T) {
	var lru Cache
	beforeLen := lru.Len()
	lru.Add("key1", String("value1"))
	afterLen := lru.Len()
	if afterLen != beforeLen+1 {
		t.Error("add key1=value1 failed")
	}

	lru.Add("key1", String("v1"))
	afterLen = lru.Len()
	if v, ok := lru.Get("key1"); !ok || afterLen != beforeLen+1 || string(v.(String)) != "v1" {
		t.Error("key exists, update failed")
	}
}

func TestGet(t *testing.T) {
	lru := New(0, nil)
	lru.Add("key1", String("value1"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "value1" {
		t.Error("hit key1=value1 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Error("miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	lru := New(int64(len(k1+k2+v1+v2)), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("k1"); ok || lru.Len() != 2 {
		t.Error("remove oldest k1 failed")
	}

	if v, ok := lru.Get("k3"); !ok || string(v.(String)) != "v3" {
		t.Error("hit k3=v3 failed")
	}
}

func TestRemove(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	lru := New(int64(len(k1+k2+k3+v1+v2+v3)), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	lru.Remove("k2")
	if _, ok := lru.Get("k2"); ok || lru.Len() != 2 {
		t.Error("remove k2 failed")
	}

	if v, ok := lru.Get("k3"); !ok || string(v.(String)) != "v3" {
		t.Error("hit k3=v3 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("v2"))
	lru.Add("k3", String("v3"))
	lru.Add("k4", String("v4"))

	expected := []string{"key1", "k2"}
	if !reflect.DeepEqual(expected, keys) {
		t.Errorf("call onEvicted failed, expected: %v, but %v\n", expected, keys)
	}
}
