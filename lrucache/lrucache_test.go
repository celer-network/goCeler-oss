// Copyright 2019-2020 Celer Network

package lrucache

import (
	"fmt"
	"testing"
)

func TestLRU(t *testing.T) {
	for i := -1; i <= 0; i++ {
		cache := NewLRUCache(i, nil)
		if cache != nil {
			t.Errorf("bad LRU size (%d) did not fail", i)
		}
	}

	dropMap := make(map[string]int)
	dropVal := func(key string, val interface{}) {
		dropMap[key] += 1
	}

	cache := NewLRUCache(64, dropVal)
	if cache == nil {
		t.Errorf("LRU failed")
	}

	for i := 0; i < 128; i++ {
		key := fmt.Sprintf("key-%d", i)
		val := fmt.Sprintf("val-%d", i)
		cache.Put(key, val)
	}

	for i := 0; i < 64; i++ {
		key := fmt.Sprintf("key-%d", i)
		if val, ok := cache.Get(key); ok {
			t.Errorf("Found %s, it was not evicted: %v", key, val)
		}
	}
	for i := 64; i < 128; i++ {
		key := fmt.Sprintf("key-%d", i)
		expVal := fmt.Sprintf("val-%d", i)
		if val, ok := cache.Get(key); !ok {
			t.Errorf("Did not find %s", key)
		} else if val != expVal {
			t.Errorf("Bad %s value: %v != %v", key, val, expVal)
		}
	}

	key := "key-64"
	expVal := "val-64"
	for i := 0; i < 5; i++ {
		if val, ok := cache.Get(key); !ok {
			t.Errorf("Did not find %s (%d)", key, i)
		} else if val != expVal {
			t.Errorf("Bad %s value (%d): %v != %v", key, i, val, expVal)
		}
	}

	if len(dropMap) != 64 {
		t.Errorf("bad length of dropMap: %d", len(dropMap))
	}
	for i := 0; i < 64; i++ {
		key = fmt.Sprintf("key-%d", i)
		if dropMap[key] != 1 {
			t.Errorf("bad dropMap count for %s: %d", key, dropMap[key])
		}
	}

	cache.Put("foo", "bar") // should evict key-65

	key = "key-65"
	if val, ok := cache.Get(key); ok {
		t.Errorf("Found %s, it was not evicted: %v", key, val)
	}

	// Modify value of most recent key.
	cache.Put("foo", "hello")
	if val, ok := cache.Get("foo"); !ok {
		t.Errorf("Did not find 'foo'")
	} else if val != "hello" {
		t.Errorf("Bad 'foo' value: %v", val)
	}

	// Modify value of an intermediate key.
	key = "key-100"
	cache.Put(key, "hello") // no evictions, key-66 still LRU
	if val, ok := cache.Get(key); !ok {
		t.Errorf("Did not find %s", key)
	} else if val != "hello" {
		t.Errorf("Bad %s value: %v", key, val)
	}

	key = "key-66"
	if val, ok := cache.Get(key); !ok {
		t.Errorf("Did not find %s", key)
	} else if val != "val-66" {
		t.Errorf("Bad %s value: %v", key, val)
	}
}
