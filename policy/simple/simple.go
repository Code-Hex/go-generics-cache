package simple

import (
	"encoding/gob"
	"fmt"
	"os"
	"sort"
	"time"
)

// Cache is a simple cache has no clear priority for evict cache.
type Cache[K comparable, V any] struct {
	items map[K]*entry[V]
}

type entry[V any] struct {
	Val       V
	CreatedAt time.Time
}

// NewCache creates a new non-thread safe cache.
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]*entry[V], 0),
	}
}

// Set sets any item to the cache. replacing any existing item.
// The default item never expires.
func (c *Cache[K, V]) Set(k K, v V) {
	c.items[k] = &entry[V]{
		Val:       v,
		CreatedAt: time.Now(),
	}
}

// Get gets an item from the cache.
// Returns the item or zero value, and a bool indicating whether the key was found.
func (c *Cache[K, V]) Get(k K) (val V, ok bool) {
	got, found := c.items[k]
	if !found {
		return
	}
	return got.Val, true
}

// Keys returns cache keys. the order is sorted by created.
func (c *Cache[K, _]) Keys() []K {
	ret := make([]K, 0, len(c.items))
	for key := range c.items {
		ret = append(ret, key)
	}
	sort.Slice(ret, func(i, j int) bool {
		i1 := c.items[ret[i]]
		i2 := c.items[ret[j]]
		return i1.CreatedAt.Before(i2.CreatedAt)
	})
	return ret
}

// Delete deletes the item with provided key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	delete(c.items, key)
}

// Saves cache state to file using gob encoder
func (c *Cache[K, V]) Save(filePath string) error {

	encodeFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Saving cache in file failed: %v", err)
	}
	defer encodeFile.Close()
	encoder := gob.NewEncoder(encodeFile)

	if err := encoder.Encode(&c.items); err != nil {
		return fmt.Errorf("Saving cache in file failed: %v", err)
	}

	return nil
}

// Loads cache state from file using gob decoder
func (c *Cache[K, V]) Load(filePath string) error {
	decodeFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening cache file failed: %v", err)
	}
	defer decodeFile.Close()

	decoder := gob.NewDecoder(decodeFile)
	// gob.Register(map[K]*entry[V]{})
	m := make(map[K]*entry[V])

	if err = decoder.Decode(&m); err != nil {
		return fmt.Errorf("loading cache from file failed: %v", err)
	}

	c.items = m

	return nil
}
