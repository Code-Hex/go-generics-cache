package simple

import (
	"sync"

	cache "github.com/Code-Hex/go-generics-cache"
)

// Cache is a simple cache has no clear priority for evict cache.
type Cache[K comparable, V any] struct {
	items map[K]*cache.Item[K, V]
	mu    sync.RWMutex
}

var _ cache.Cache[interface{}, any] = (*Cache[interface{}, any])(nil)

// NewCache creates a new cache.
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]*cache.Item[K, V]),
	}
}

// Set sets any item to the cache. replacing any existing item.
// The default item never expires.
func (c *Cache[K, V]) Set(k K, v V, opts ...cache.ItemOption) {
	c.mu.Lock()
	c.items[k] = cache.NewItem(k, v, opts...)
	c.mu.Unlock()
}

// Get gets an item from the cache.
// Returns the item or zero value, and a bool indicating whether the key was found.
func (c *Cache[K, V]) Get(k K) (val V, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	got, found := c.items[k]
	if !found {
		return
	}
	if got.HasExpired() {
		return
	}
	return got.Value, true
}

// Keys returns cache keys. the order is random.
func (c *Cache[K, _]) Keys() []K {
	ret := make([]K, 0, len(c.items))
	for key := range c.items {
		ret = append(ret, key)
	}
	return ret
}
