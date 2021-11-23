package lru

import (
	"container/list"
	"sync"
)

// Cache is a thread safe LRU cache
type Cache[K comparable, V any] struct {
	cap   int
	list  *list.List
	items map[K]*list.Element
	mu    sync.RWMutex
}

type item[K comparable, V any] struct {
	Key   K
	Value V
}

// New creates a new LRU cache whose capacity is the default size (128).
func New[K comparable, V any]() *Cache[K, V] {
	return NewCap[K, V](128)
}

// NewCap creates a new LRU cache whose capacity is the specified size.
func NewCap[K comparable, V any](cap int) *Cache[K, V] {
	return &Cache[K, V]{
		cap:   cap,
		list:  list.New(),
		items: make(map[K]*list.Element),
	}
}

// Get looks up a key's value from the cache.
func (c *Cache[K, V]) Get(key K) (zero V, _ bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if e, ok := c.items[key]; ok {
		// updates cache order
		c.list.MoveToFront(e)
		return e.Value.(*item[K, V]).Value, true
	}
	return
}

// Set sets a value to the cache with key. replacing any existing value.
func (c *Cache[K, V]) Set(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.items[key]; ok {
		// updates cache order
		c.list.MoveToFront(e)
		e.Value.(*item[K, V]).Value = val
		return
	}

	e := c.list.PushFront(&item[K, V]{
		Key:   key,
		Value: val,
	})
	c.items[key] = e

	if c.list.Len() > c.cap {
		c.deleteOldest()
	}
}

// Keys returns the keys of the cache. the order is from oldest to newest.
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]K, 0, len(c.items))
	for ent := c.list.Back(); ent != nil; ent = ent.Prev() {
		keys = append(keys, ent.Value.(*item[K, V]).Key)
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list.Len()
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.items[key]; ok {
		c.delete(e)
	}
}

// Contains reports whether key is within cache.
func (c *Cache[K, V]) Contains(key K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.items[key]
	return ok
}

func (c *Cache[K, V]) deleteOldest() {
	e := c.list.Back()
	c.delete(e)
}

func (c *Cache[K, V]) delete(e *list.Element) {
	c.list.Remove(e)
	item := e.Value.(*item[K, V])
	delete(c.items, item.Key)
}
