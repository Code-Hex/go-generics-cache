package lfu

import (
	"container/heap"
	"sync"

	cache "github.com/Code-Hex/go-generics-cache"
)

// Cache is a thread safe LRU cache
type Cache[K comparable, V any] struct {
	cap   int
	queue *priorityQueue[K, V]
	items map[K]*entry[K, V]
	mu    sync.RWMutex
}

var _ cache.Cache[interface{}, any] = (*Cache[interface{}, any])(nil)

// NewCache creates a new LFU cache whose capacity is the default size (128).
func NewCache[K comparable, V any]() *Cache[K, V] {
	return NewCacheWithCap[K, V](128)
}

// NewCacheWithCap creates a new LFU cache whose capacity is the specified size.
func NewCacheWithCap[K comparable, V any](cap int) *Cache[K, V] {
	return &Cache[K, V]{
		cap:   cap,
		queue: newPriorityQueue[K, V](cap),
		items: make(map[K]*entry[K, V], cap),
	}
}

// Get looks up a key's value from the cache.
func (c *Cache[K, V]) Get(key K) (zero V, _ bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.items[key]
	if !ok {
		return
	}
	if e.item.HasExpired() {
		return
	}
	e.item.Referenced()
	heap.Fix(c.queue, e.index)
	return e.item.Value, true
}

// Set sets a value to the cache with key. replacing any existing value.
func (c *Cache[K, V]) Set(key K, val V, opts ...cache.ItemOption) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.items[key]; ok {
		c.queue.update(e, val)
		return
	}

	e := newEntry(key, val, opts...)
	heap.Push(c.queue, e)
	c.items[key] = e

	if len(c.items) > c.cap {
		evictedEntry := heap.Pop(c.queue).(*entry[K, V])
		delete(c.items, evictedEntry.item.Key)
	}
}

// Keys returns the keys of the cache. the order is from oldest to newest.
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]K, 0, len(c.items))
	for _, entry := range *c.queue {
		keys = append(keys, entry.item.Key)
	}
	return keys
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.items[key]; ok {
		heap.Remove(c.queue, e.index)
		delete(c.items, key)
	}
}

// Contains reports whether key is within cache.
func (c *Cache[K, V]) Contains(key K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok {
		return false
	}
	return !e.item.HasExpired()
}
