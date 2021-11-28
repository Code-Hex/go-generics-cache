package lru

import (
	"container/list"
	"sync"

	cache "github.com/Code-Hex/go-generics-cache"
)

// Cache is a thread safe LRU cache
type Cache[K comparable, V any] struct {
	cap   int
	list  *list.List
	items map[K]*list.Element
	mu    sync.RWMutex
}

var _ cache.Cache[interface{}, any] = (*Cache[interface{}, any])(nil)

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
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok {
		return
	}
	item := e.Value.(*cache.Item[K, V])
	if item.HasExpired() {
		return
	}
	// updates cache order
	c.list.MoveToFront(e)
	return item.Value, true
}

// Set sets a value to the cache with key. replacing any existing value.
func (c *Cache[K, V]) Set(key K, val V, opts ...cache.ItemOption) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.items[key]; ok {
		// updates cache order
		c.list.MoveToFront(e)
		e.Value.(*cache.Item[K, V]).Value = val
		return
	}

	item := cache.NewItem(key, val, opts...)
	e := c.list.PushFront(item)
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
		item := ent.Value.(*cache.Item[K, V])
		keys = append(keys, item.Key)
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
	e, ok := c.items[key]
	if !ok {
		return false
	}
	item := e.Value.(*cache.Item[K, V])
	return !item.HasExpired()
}

func (c *Cache[K, V]) deleteOldest() {
	e := c.list.Back()
	c.delete(e)
}

func (c *Cache[K, V]) delete(e *list.Element) {
	c.list.Remove(e)
	item := e.Value.(*cache.Item[K, V])
	delete(c.items, item.Key)
}
