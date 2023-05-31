package fifo

import (
	"container/list"
	"encoding/gob"
	"fmt"
	"os"
)

// Cache is used a FIFO (First in first out) cache replacement policy.
//
// In FIFO the item that enter the cache first is evicted first
// w/o any regard of how often or how many times it was accessed before.
type Cache[K comparable, V any] struct {
	items    map[K]*list.Element
	queue    *list.List // keys
	capacity int
}

type entry[K comparable, V any] struct {
	key K
	val V
}

// Option is an option for FIFO cache.
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

// NewCache creates a new non-thread safe FIFO cache whose capacity is the default size (128).
func NewCache[K comparable, V any](opts ...Option) *Cache[K, V] {
	o := newOptions()
	for _, optFunc := range opts {
		optFunc(o)
	}
	return &Cache[K, V]{
		items:    make(map[K]*list.Element, o.capacity),
		queue:    list.New(),
		capacity: o.capacity,
	}
}

// Set sets any item to the cache. replacing any existing item.
func (c *Cache[K, V]) Set(key K, val V) {
	if c.queue.Len() == c.capacity {
		e := c.dequeue()
		delete(c.items, e.Value.(*entry[K, V]).key)
	}
	c.Delete(key) // delete old key if already exists specified key.
	entry := &entry[K, V]{
		key: key,
		val: val,
	}
	e := c.queue.PushBack(entry)
	c.items[key] = e
}

// Get gets an item from the cache.
// Returns the item or zero value, and a bool indicating whether the key was found.
func (c *Cache[K, V]) Get(k K) (val V, ok bool) {
	got, found := c.items[k]
	if !found {
		return
	}
	return got.Value.(*entry[K, V]).val, true
}

// Keys returns cache keys.
func (c *Cache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))
	for e := c.queue.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry[K, V]).key)
	}
	return keys
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	if e, ok := c.items[key]; ok {
		c.queue.Remove(e)
		delete(c.items, key)
	}
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	return c.queue.Len()
}

func (c *Cache[K, V]) dequeue() *list.Element {
	e := c.queue.Front()
	c.queue.Remove(e)
	return e
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
	m := make(map[K]*list.Element)

	if err := decoder.Decode(&m); err != nil {
		return fmt.Errorf("loading cache from file failed: %v", err)
	}

	c.items = m

	return nil
}
