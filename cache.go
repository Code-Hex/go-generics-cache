package cache

import (
	"sync"
	"time"

	"github.com/Code-Hex/go-generics-cache/lfu"
	"github.com/Code-Hex/go-generics-cache/lru"
	"github.com/Code-Hex/go-generics-cache/simple"
)

// Interface is a common-cache interface.
type Interface[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, val V)
	Keys() []K
	Delete(key K)
}

var (
	_ Interface[any, any] = (*simple.Cache[any, any])(nil)
	_ Interface[any, any] = (*lru.Cache[any, any])(nil)
	_ Interface[any, any] = (*lfu.Cache[any, any])(nil)
)

// Item is an item
type Item[K comparable, V any] struct {
	Key        K
	Value      V
	Expiration time.Duration
}

var nowFunc = time.Now

// ItemOption is an option for cache item.
type ItemOption func(*itemOptions)

type itemOptions struct {
	expiration time.Duration // default none
}

// WithExpiration is an option to set expiration time for any items.
// If the expiration is zero or negative value, it treat as no expiration.
func WithExpiration(exp time.Duration) ItemOption {
	return func(o *itemOptions) {
		o.expiration = exp
	}
}

// newItem creates a new item with specified any options.
func newItem[K comparable, V any](key K, val V, opts ...ItemOption) *Item[K, V] {
	o := new(itemOptions)
	for _, optFunc := range opts {
		optFunc(o)
	}
	return &Item[K, V]{
		Key:        key,
		Value:      val,
		Expiration: o.expiration,
	}
}

// Cache is a cache.
type Cache[K comparable, V any] struct {
	cache       Interface[K, *Item[K, V]]
	expirations map[K]chan struct{}
	// mu is used to do lock in some method process.
	mu sync.RWMutex
}

// Option is an option for cache.
type Option[K comparable, V any] func(*options[K, V])

type options[K comparable, V any] struct {
	cache Interface[K, *Item[K, V]]
}

func newOptions[K comparable, V any]() *options[K, V] {
	return &options[K, V]{
		cache: simple.NewCache[K, *Item[K, V]](),
	}
}

// AsLRU is an option to make a new Cache as LRU algorithm.
func AsLRU[K comparable, V any](cap int) Option[K, V] {
	return func(o *options[K, V]) {
		o.cache = lru.NewCacheWithCap[K, *Item[K, V]](cap)
	}
}

// AsLFU is an option to make a new Cache as LFU algorithm.
func AsLFU[K comparable, V any](cap int) Option[K, V] {
	return func(o *options[K, V]) {
		o.cache = lfu.NewCacheWithCap[K, *Item[K, V]](cap)
	}
}

// New creates a new Cache.
func New[K comparable, V any](opts ...Option[K, V]) *Cache[K, V] {
	o := newOptions[K, V]()
	for _, optFunc := range opts {
		optFunc(o)
	}
	return &Cache[K, V]{
		cache:       o.cache,
		expirations: make(map[K]chan struct{}, 0),
	}
}

// Get looks up a key's value from the cache.
func (c *Cache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.cache.Get(key)
	if !ok {
		return
	}
	return item.Value, true
}

// Set sets a value to the cache with key. replacing any existing value.
func (c *Cache[K, V]) Set(key K, val V, opts ...ItemOption) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item := newItem(key, val, opts...)
	if item.Expiration <= 0 {
		c.cache.Set(key, item)
		return
	}

	if _, ok := c.cache.Get(key); ok {
		c.doneWatchExpiration(key)
	}

	c.cache.Set(key, item)
	c.installExpirationWatcher(item.Key, item.Expiration)
}

func (c *Cache[K, V]) installExpirationWatcher(key K, exp time.Duration) {
	done := make(chan struct{})
	c.expirations[key] = done
	go func() {
		select {
		case <-time.After(exp):
			c.Delete(key)
		case <-done:
		}
	}()
}

func (c *Cache[K, V]) doneWatchExpiration(key K) {
	if ch, ok := c.expirations[key]; ok {
		close(ch)
	}
}

// Keys returns the keys of the cache. the order is relied on algorithms.
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache.Keys()
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Delete(key)
}

// Contains reports whether key is within cache.
func (c *Cache[K, V]) Contains(key K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.cache.Get(key)
	return ok
}

// NumberCache is a in-memory cache which is able to store only Number constraint.
type NumberCache[K comparable, V Number] struct {
	*Cache[K, V]
	// nmu is used to do lock in Increment/Decrement process.
	// Note that this must be here as a separate mutex because mu in Cache struct is Locked in Get,
	// and if we call mu.Lock in Increment/Decrement, it will cause deadlock.
	nmu sync.Mutex
}

// NewNumber creates a new cache for Number constraint.
func NewNumber[K comparable, V Number](opts ...Option[K, V]) *NumberCache[K, V] {
	return &NumberCache[K, V]{
		Cache: New(opts...),
	}
}

// Increment an item of type Number constraint by n.
// Returns the incremented value.
func (nc *NumberCache[K, V]) Increment(key K, n V) V {
	// In order to avoid lost update, we must lock whole Increment/Decrement process.
	nc.nmu.Lock()
	defer nc.nmu.Unlock()
	got, _ := nc.Cache.Get(key)
	nv := got + n
	nc.Cache.Set(key, nv)
	return nv
}

// Decrement an item of type Number constraint by n.
// Returns the decremented value.
func (nc *NumberCache[K, V]) Decrement(key K, n V) V {
	nc.nmu.Lock()
	defer nc.nmu.Unlock()
	got, _ := nc.Cache.Get(key)
	nv := got - n
	nc.Cache.Set(key, nv)
	return nv
}
