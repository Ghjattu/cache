package cache

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(string) ([]byte, error)

// Get implements Getter interface function.
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Group is a cache namespace and associated data loaded spread over
// a group of 1 or more machines.
type Group struct {
	name      string // Each group has a unique name.
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup creates a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) (*Group, error) {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()

	// Check if the name is already associated with any group.
	if _, exist := groups[name]; exist {
		return nil, fmt.Errorf("name %s already exists", name)
	}

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g

	return g, nil
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	g := groups[name]
	return g
}

// Get gets value for a key from the cache.
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}

	// cache hit, return the data
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[cache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

// getLocally calls getter.get function to get data, and adds it to cache.
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	v := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, v)
	return v, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
