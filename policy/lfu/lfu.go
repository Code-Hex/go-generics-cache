package lfu

import (
	"container/heap"
	"encoding/gob"
	"fmt"
	"os"
)

// Cache is used a LFU (Least-frequently used) cache replacement policy.
//
// Counts how often an item is needed. Those that are used least often are discarded first.
// This works very similar to LRU except that instead of storing the value of how recently
// a block was accessed, we store the value of how many times it was accessed. So of course
// while running an access sequence we will replace a block which was used fewest times from our cache.
type Cache[K comparable, V any] struct {
	cap   int
	queue *priorityQueue[K, V]
	items map[K]*entry[K, V]
}

// Option is an option for LFU cache.
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

// NewCache creates a new non-thread safe LFU cache whose capacity is the default size (128).
func NewCache[K comparable, V any](opts ...Option) *Cache[K, V] {
	o := newOptions()
	for _, optFunc := range opts {
		optFunc(o)
	}
	return &Cache[K, V]{
		cap:   o.capacity,
		queue: newPriorityQueue[K, V](o.capacity),
		items: make(map[K]*entry[K, V], o.capacity),
	}
}

// Get looks up a key's value from the cache.
func (c *Cache[K, V]) Get(key K) (zero V, _ bool) {
	e, ok := c.items[key]
	if !ok {
		return
	}
	e.referenced()
	heap.Fix(c.queue, e.Index)
	return e.Val, true
}

// Set sets a value to the cache with key. replacing any existing value.
//
// If value satisfies "interface{ GetReferenceCount() int }", the value of
// the GetReferenceCount() method is used to set the initial value of reference count.
func (c *Cache[K, V]) Set(key K, val V) {
	if e, ok := c.items[key]; ok {
		c.queue.update(e, val)
		return
	}

	if len(c.items) == c.cap {
		evictedEntry := heap.Pop(c.queue).(*entry[K, V])
		delete(c.items, evictedEntry.Key)
	}

	e := newEntry(key, val)
	heap.Push(c.queue, e)
	c.items[key] = e
}

// Keys returns the keys of the cache. the order is from oldest to newest.
func (c *Cache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))
	for _, entry := range *c.queue {
		keys = append(keys, entry.Key)
	}
	return keys
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	if e, ok := c.items[key]; ok {
		heap.Remove(c.queue, e.Index)
		delete(c.items, key)
	}
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	return c.queue.Len()
}

// Saves cache state to file using gob encoder
func (c *Cache[K, V]) Save(filePath string) error {
	encodeFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Saving cache in file failed: %v", err)
	}
	defer encodeFile.Close()
	encoder := gob.NewEncoder(encodeFile)

	if err := encoder.Encode(c.items); err != nil {
		return fmt.Errorf("Saving cache in file failed: %v", err)
	}

	return nil
}

// Loads cache state from file using gob decoder
func (c *Cache[K, V]) Load(filePath string) error {
	decodeFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("loading cache from file failed: %v", err)
	}
	defer decodeFile.Close()

	decoder := gob.NewDecoder(decodeFile)
	m := make(map[K]*entry[K, V])

	if err := decoder.Decode(&m); err != nil {
		return fmt.Errorf("loading cache from file failed: %v", err)
	}

	c.items = m

	return nil
}
