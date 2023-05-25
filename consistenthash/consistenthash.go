package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash function maps byte to uint32.
type Hash func(data []byte) uint32

type Map struct {
	hash Hash
	keys []int

	// replicas is the number of virtual nodes
	// corresponding to a real node.
	replicas int

	// hashMap maps virtual node to real node,
	// key is virtual node's hash value, value is real node's name.
	hashMap map[int]string
}

// New creates a new instance of Map.
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// IsEmpty returns true if there are no items available.
func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closest item from the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
