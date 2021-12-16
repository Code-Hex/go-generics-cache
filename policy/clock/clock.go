package clock

import (
	"container/ring"
)

// Cache is used The clock cache replacement policy.
//
// The clock algorithm keeps a circular list of pages in memory, with
// the "hand" (iterator) pointing to the last examined page frame in the list.
// When a page fault occurs and no empty frames exist, then the R (referenced) bit
// is inspected at the hand's location. If R is 0, the new page is put in place of
// the page the "hand" points to, and the hand is advanced one position. Otherwise,
// the R bit is cleared, then the clock hand is incremented and the process is
// repeated until a page is replaced.
type Cache[K comparable, V any] struct {
	items    map[K]*ring.Ring
	hand     *ring.Ring
	head     *ring.Ring
	capacity int
}

type entry[K comparable, V any] struct {
	key            K
	val            V
	referenceCount int
}

// Option is an option for clock cache.
type Option func(*options)

type options struct {
	capacity int
}

func newOptions() *options {
	return &options{
		capacity: 128,
	}
}

// WithCapacity is an option to set cache capacity.
func WithCapacity(cap int) Option {
	return func(o *options) {
		o.capacity = cap
	}
}

// NewCache creates a new non-thread safe clock cache whose capacity is the default size (128).
func NewCache[K comparable, V any](opts ...Option) *Cache[K, V] {
	o := newOptions()
	for _, optFunc := range opts {
		optFunc(o)
	}
	r := ring.New(o.capacity)
	return &Cache[K, V]{
		items:    make(map[K]*ring.Ring, o.capacity),
		hand:     r,
		head:     r,
		capacity: o.capacity,
	}
}

// Set sets any item to the cache. replacing any existing item.
func (c *Cache[K, V]) Set(key K, val V) {
	if e, ok := c.items[key]; ok {
		entry := e.Value.(*entry[K, V])
		entry.referenceCount++
		entry.val = val
		return
	}
	c.evict()
	c.hand.Value = &entry[K, V]{
		key:            key,
		val:            val,
		referenceCount: 1,
	}
	c.items[key] = c.hand
	c.hand = c.hand.Next()
}

// Get looks up a key's value from the cache.
func (c *Cache[K, V]) Get(key K) (zero V, _ bool) {
	e, ok := c.items[key]
	if !ok {
		return
	}
	entry := e.Value.(*entry[K, V])
	entry.referenceCount = 1
	return entry.val, true
}

func (c *Cache[K, V]) evict() {
	if c.hand.Value == nil {
		return
	}
	for c.hand.Value.(*entry[K, V]).referenceCount > 0 {
		c.hand.Value.(*entry[K, V]).referenceCount--
		c.hand = c.hand.Next()
	}
	entry := c.hand.Value.(*entry[K, V])
	delete(c.items, entry.key)
	c.hand.Value = nil
}

// Keys returns the keys of the cache. the order as same as current ring order.
func (c *Cache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))
	r := c.head
	if r.Value == nil {
		return []K{}
	}
	// the first element
	keys = append(keys, r.Value.(*entry[K, V]).key)

	// iterating
	for p := c.head.Next(); p != r; p = p.Next() {
		if p.Value == nil {
			break
		}
		e := p.Value.(*entry[K, V])
		if e.referenceCount > 0 {
			keys = append(keys, e.key)
		}
	}
	return keys
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	if e, ok := c.items[key]; ok {
		delete(c.items, key)
		e.Value.(*entry[K, V]).referenceCount = 0
	}
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	return len(c.items)
}
