package cache

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrNotFound is an error which indicate an item is not found.
	ErrNotFound = errors.New("not found item")

	// ErrExpired is an error which indicate an item is expired.
	ErrExpired = errors.New("expired item")
)

// Item is an item
type Item[T any] struct {
	Value      T
	Expiration time.Duration
	CreatedAt  time.Time
}

var nowFunc = time.Now

// HasExpired returns true if the item has expired.
// If the item's expiration is zero value, returns false.
func (i Item[T]) HasExpired() bool {
	if i.Expiration <= 0 {
		return false
	}
	return i.CreatedAt.Add(i.Expiration).Before(nowFunc())
}

// Cache is a base struct for creating in-memory cache.
type Cache[K comparable, V any] struct {
	items map[K]Item[V]
	mu    sync.RWMutex
}

// New creates a new cache.
func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]Item[V]),
	}
}

// ItemOption is an option for cache item.
type ItemOption func(o *options)

type options struct {
	expiration time.Duration // default none
}

// WithExpiration is an option to set expiration time for any items.
func WithExpiration(exp time.Duration) ItemOption {
	return func(o *options) {
		o.expiration = exp
	}
}

// Set sets any item to the cache. replacing any existing item.
// The default item never expires.
func (c *Cache[K, V]) Set(k K, v V, opts ...ItemOption) {
	o := new(options)
	for _, optFunc := range opts {
		optFunc(o)
	}
	item := Item[V]{
		Value:      v,
		Expiration: o.expiration,
		CreatedAt:  nowFunc(),
	}
	c.SetItem(k, item)
}

// SetItem sets any item to the cache. replacing any existing item.
// The default item never expires.
func (c *Cache[K, V]) SetItem(k K, v Item[V]) {
	c.mu.Lock()
	c.items[k] = v
	c.mu.Unlock()
}

// Get gets an item from the cache.
// Returns the item or zero value, and a bool indicating whether the key was found.
func (c *Cache[K, V]) Get(k K) (val V, ok bool) {
	item, err := c.GetItem(k)
	if err != nil {
		return
	}
	return item.Value, true
}

// GetItem gets an item from the cache.
// Returns an error if the item was not found or expired. If there is no error, the
// incremented value is returned.
func (c *Cache[K, V]) GetItem(k K) (val Item[V], _ error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	got, found := c.items[k]
	if !found {
		return val, fmt.Errorf("key[%v]: %w", k, ErrNotFound)
	}
	if got.HasExpired() {
		return val, fmt.Errorf("key[%v]: %w", k, ErrExpired)
	}
	return got, nil
}

// Keys returns cache keys. the order is random.
func (c *Cache[K, _]) Keys() []K {
	ret := make([]K, 0, len(c.items))
	for key := range c.items {
		ret = append(ret, key)
	}
	return ret
}

// NumberCache is a in-memory cache which is able to store only Number constraint.
type NumberCache[K comparable, V Number] struct {
	*Cache[K, V]
}

// NewNumber creates a new cache for Number constraint.
func NewNumber[K comparable, V Number]() *NumberCache[K, V] {
	return &NumberCache[K, V]{
		Cache: New[K, V](),
	}
}

// Increment an item of type Number constraint by n.
// Returns an error if the item was not found or expired. If there is no error, the
// incremented value is returned.
func (nc *NumberCache[K, V]) Increment(k K, n V) (val V, err error) {
	got, err := nc.Cache.GetItem(k)
	if err != nil {
		return val, err
	}

	nv := got.Value + n
	got.Value = nv
	nc.Cache.SetItem(k, got)

	return nv, nil
}

// Decrement an item of type Number constraint by n.
// Returns an error if the item was not found or expired. If there is no error, the
// decremented value is returned.
func (nc *NumberCache[K, V]) Decrement(k K, n V) (val V, err error) {
	got, err := nc.Cache.GetItem(k)
	if err != nil {
		return val, err
	}

	nv := got.Value - n
	got.Value = nv
	nc.Cache.SetItem(k, got)

	return nv, nil
}
