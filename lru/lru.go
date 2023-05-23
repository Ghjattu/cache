package lru

import (
	"container/list"
)

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	// maximum memory can be used
	maxBytes int64
	// memory used currently
	nBytes int64

	ll    *list.List
	cache map[string]*list.Element

	// optional callback function and executed when an entry is purged from the cache
	OnEvicted func(key string, value Value)
}

// New creates a new Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nBytes:    0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add adds a value to the cache
func (c *Cache) Add(key string, value Value) {
	if c.cache == nil {
		c.cache = make(map[string]*list.Element)
		c.ll = list.New()
	}

	if element, hit := c.cache[key]; hit { // key already exists, update value
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else { // key does not exist, add an item
		element := c.ll.PushFront(&entry{key, value})
		c.cache[key] = element
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache
func (c *Cache) Get(key string) (value Value, ok bool) {
	if c.cache == nil {
		return nil, false
	}

	if element, hit := c.cache[key]; hit {
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)

		return kv.value, true
	}

	return nil, false
}

// Remove removes the provided key from the cache
func (c *Cache) Remove(key string) {
	if c.cache == nil {
		return
	}

	if element, hit := c.cache[key]; hit {
		c.removeElement(element)
	}
}

// RemoveOldest removes the oldest item from cache
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}

	element := c.ll.Back()
	if element != nil {
		c.removeElement(element)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())

	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}

	return c.ll.Len()
}
