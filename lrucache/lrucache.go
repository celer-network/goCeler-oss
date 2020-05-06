// Copyright 2019-2020 Celer Network
//
// Manage a fixed-size LRU cache of key/value pairs where the keys are
// strings and the values are generic objects.  Maintain a free-list of
// entries to reduce GC overhead.  Invoke an optional callback when a
// value is dropped to allow the application to destroy that object.

package lrucache

import (
	"sync"
)

// LRUCache is a fixed-size cache of key/value pairs.  The lookup map
// is used to access values by their key.  The LRU doubly-linked list
// is used to track entries from the least-recently used (the head) to
// the most recently used (the tail).  An optional callback notifies
// the application when a value is dropped so it can be destroyed.
type LRUCache struct {
	mu       sync.Mutex
	lookup   map[string]*entry
	lruHead  *entry
	lruTail  *entry
	freeList *entry
	dropVal  DropValueCallback
}

type entry struct {
	prev *entry
	next *entry
	key  string
	val  interface{}
}

type DropValueCallback func(key string, val interface{})

// NewLRUCache creates an LRU cache of a given size.  An optional
// callback is invoked when a value is dropped from the cache,
// either overwritten or evicted.
func NewLRUCache(size int, dropVal DropValueCallback) *LRUCache {
	if size <= 0 {
		return nil
	}

	entries := make([]entry, size)
	for i := 0; i < size-1; i++ {
		entries[i].next = &entries[i+1]
	}
	entries[size-1].next = nil

	cache := &LRUCache{
		lookup:   make(map[string]*entry, size),
		freeList: &entries[0],
		dropVal:  dropVal,
	}
	return cache
}

// Put stores a key/value pair into the cache.  The ownership of the
// value object is transferred to the cache, the application should
// not modify or reuse that object.
func (c *LRUCache) Put(key string, val interface{}) {
	hasDrop := false
	var dropKey string
	var dropVal interface{}

	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.lookup[key]
	if ok {
		if val != e.val {
			// New value replaces an old one.
			hasDrop = true
			dropKey, dropVal = e.key, e.val
			e.val = val
		}
		c.moveToLRUTail(e)
	} else {
		if c.freeList != nil {
			// Grab an entry from the free-list.
			e = c.freeList
			c.freeList = e.next
		} else {
			// Empty free-list: evict an entry from the
			// head of the list (least-recently used).
			e = c.lruHead
			c.lruHead = e.next
			if e.next != nil {
				e.next.prev = nil
			}
			hasDrop = true
			dropKey, dropVal = e.key, e.val
			delete(c.lookup, dropKey)
		}

		// Store the new key/value pair.
		e.key, e.val = key, val
		if c.lruHead == nil {
			c.lruHead = e
		}
		c.lruAppend(e)
		c.lookup[key] = e
	}

	// Notify the application if a value was dropped.
	if hasDrop && c.dropVal != nil {
		c.dropVal(dropKey, dropVal)
	}
}

// Get looks up the value in the cache for a given key.  If found,
// it returns (value, true).  Otherwise it returns (nil, false).
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.lookup[key]
	if !ok {
		return nil, false
	}

	c.moveToLRUTail(e)
	return e.val, true
}

// Move the given entry from its current position to the tail of
// the LRU list (i.e. make it most-recently used).
func (c *LRUCache) moveToLRUTail(e *entry) {
	if c.lruTail != e {
		if c.lruHead == e {
			c.lruHead = e.next
		} else {
			e.prev.next = e.next
		}
		if e.next != nil {
			e.next.prev = e.prev
		}
		c.lruAppend(e)
	}
}

// Append the given entry to the tail of the LRU list.
func (c *LRUCache) lruAppend(e *entry) {
	e.next = nil
	e.prev = c.lruTail
	if e.prev != nil {
		e.prev.next = e
	}
	c.lruTail = e
}
