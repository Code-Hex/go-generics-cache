package mru

import (
	"container/list"
	"encoding/gob"
	"fmt"
	"os"
)

// Cache is used a MRU (Most recently used) cache replacement policy.
//
// In contrast to Least Recently Used (LRU), MRU discards the most recently used items first.
type Cache[K comparable, V any] struct {
	Cap   int
	List  *list.List
	Items map[K]*list.Element
}

type entry[K comparable, V any] struct {
	Key K
	Val V
}

// Option is an option for MRU cache.
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

// NewCache creates a new non-thread safe MRU cache whose capacity is the default size (128).
func NewCache[K comparable, V any](opts ...Option) *Cache[K, V] {
	o := newOptions()
	for _, optFunc := range opts {
		optFunc(o)
	}
	return &Cache[K, V]{
		Cap:   o.capacity,
		List:  list.New(),
		Items: make(map[K]*list.Element, o.capacity),
	}
}

// Get looks up a key's value from the cache.
func (c *Cache[K, V]) Get(key K) (zero V, _ bool) {
	e, ok := c.Items[key]
	if !ok {
		return
	}
	// updates cache order
	c.List.MoveToBack(e)
	return e.Value.(*entry[K, V]).Val, true
}

// Set sets a value to the cache with key. replacing any existing value.
func (c *Cache[K, V]) Set(key K, val V) {
	if e, ok := c.Items[key]; ok {
		// updates cache order
		c.List.MoveToBack(e)
		entry := e.Value.(*entry[K, V])
		entry.Val = val
		return
	}

	if c.List.Len() == c.Cap {
		c.deleteNewest()
	}

	newEntry := &entry[K, V]{
		Key: key,
		Val: val,
	}
	e := c.List.PushBack(newEntry)
	c.Items[key] = e
}

// Keys returns the keys of the cache. the order is from recently used.
func (c *Cache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.Items))
	for ent := c.List.Back(); ent != nil; ent = ent.Prev() {
		entry := ent.Value.(*entry[K, V])
		keys = append(keys, entry.Key)
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	return c.List.Len()
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	if e, ok := c.Items[key]; ok {
		c.delete(e)
	}
}

func (c *Cache[K, V]) deleteNewest() {
	e := c.List.Front()
	c.delete(e)
}

func (c *Cache[K, V]) delete(e *list.Element) {
	c.List.Remove(e)
	entry := e.Value.(*entry[K, V])
	delete(c.Items, entry.Key)
}

// Saves cache state to file using gob encoder
func (c *Cache[K, V]) Save(filePath string) error {
	encodeFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Saving cache in file failed: %v", err)
	}
	defer encodeFile.Close()
	encoder := gob.NewEncoder(encodeFile)

	if err := encoder.Encode(&c.Items); err != nil {
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
	// var m *Cache[K, V]
	m := make(map[K]*list.Element, c.Cap)
	if err := decoder.Decode(&m); err != nil {
		return fmt.Errorf("loading cache from file failed: %v", err)
	}

	// c.Items = m.Items
	// c.Cap = m.Cap
	// c.List = m.List

	println("loaded items count = ", len(c.Items))

	return nil
}
