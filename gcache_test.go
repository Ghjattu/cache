package cache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var slowDB = map[string]string{
	"Alice":     "20",
	"Bob":       "21",
	"Charlotte": "22",
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(s string) ([]byte, error) {
		return []byte(s), nil
	})

	expected := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(expected, v) {
		t.Error("get callback function failed")
	}
}

func TestGet(t *testing.T) {
	// If cache does not hit, retrieve data from slowDB.
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		log.Println("[slowDB] search key: ", key)
		if v, ok := slowDB[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s does not exist", key)
	})

	group := NewGroup("age", 2<<10, f)
	for k, v := range slowDB {
		// This group.Get will not hit cache.
		if view, err := group.Get(k); err != nil || string(view.b) != v {
			t.Fatalf("retrieve key=%s data failed", k)
		}
		// This group.Get will hit cache.
		if view, err := group.Get(k); err != nil || string(view.b) != v {
			t.Fatalf("cache for key=%s miss", k)
		}
	}

	if view, err := group.Get("unknown"); err == nil {
		t.Fatalf("the value of key=unknown should be empty, but %s got", string(view.b))
	}
}
