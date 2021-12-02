package lru

import (
	"container/list"
)

// Cache is a thread safe LRU cache
type Cache[K comparable, V any] struct {
	cap   int
	list  *list.List
	items map[K]*list.Element
}

type entry[K comparable, V any] struct {
	key K
	val V
}

// NewCache creates a new LRU cache whose capacity is the default size (128).
func NewCache[K comparable, V any]() *Cache[K, V] {
	return NewCacheWithCap[K, V](128)
}

// NewCacheWithCap creates a new LRU cache whose capacity is the specified size.
func NewCacheWithCap[K comparable, V any](cap int) *Cache[K, V] {
	return &Cache[K, V]{
		cap:   cap,
		list:  list.New(),
		items: make(map[K]*list.Element, cap),
	}
}

// Get looks up a key's value from the cache.
func (c *Cache[K, V]) Get(key K) (zero V, _ bool) {
	e, ok := c.items[key]
	if !ok {
		return
	}
	// updates cache order
	c.list.MoveToFront(e)
	return e.Value.(*entry[K, V]).val, true
}

// Set sets a value to the cache with key. replacing any existing value.
func (c *Cache[K, V]) Set(key K, val V) {
	if e, ok := c.items[key]; ok {
		// updates cache order
		c.list.MoveToFront(e)
		entry := e.Value.(*entry[K, V])
		entry.val = val
		return
	}

	newEntry := &entry[K, V]{
		key: key,
		val: val,
	}
	e := c.list.PushFront(newEntry)
	c.items[key] = e

	if c.list.Len() > c.cap {
		c.deleteOldest()
	}
}

// Keys returns the keys of the cache. the order is from oldest to newest.
func (c *Cache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))
	for ent := c.list.Back(); ent != nil; ent = ent.Prev() {
		entry := ent.Value.(*entry[K, V])
		keys = append(keys, entry.key)
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	return c.list.Len()
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	if e, ok := c.items[key]; ok {
		c.delete(e)
	}
}

func (c *Cache[K, V]) deleteOldest() {
	e := c.list.Back()
	c.delete(e)
}

func (c *Cache[K, V]) delete(e *list.Element) {
	c.list.Remove(e)
	entry := e.Value.(*entry[K, V])
	delete(c.items, entry.key)
}
