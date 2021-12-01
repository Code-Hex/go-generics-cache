package cache

import (
	"sync"
	"time"
)

// Cache is a common-cache interface.
type Cache[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, val V, opts ...ItemOption)
	Keys() []K
	Delete(key K)
	Contains(key K) bool
}

// Item is an item
type Item[K comparable, V any] struct {
	Key            K
	Value          V
	ReferenceCount int
	Expiration     time.Duration
	CreatedAt      time.Time
	ReferencedAt   time.Time
}

// Referenced increments a reference counter and updates `ReferencedAt`
// to current time.
func (i *Item[K, V]) Referenced() {
	i.ReferenceCount++
	i.ReferencedAt = nowFunc()
}

var nowFunc = time.Now

// HasExpired returns true if the item has expired.
// If the item's expiration is zero value, returns false.
func (i Item[K, T]) HasExpired() bool {
	if i.Expiration <= 0 {
		return false
	}
	return i.CreatedAt.Add(i.Expiration).Before(nowFunc())
}

// ItemOption is an option for cache item.
type ItemOption func(*itemOptions)

type itemOptions struct {
	expiration time.Duration // default none
}

// WithExpiration is an option to set expiration time for any items.
func WithExpiration(exp time.Duration) ItemOption {
	return func(o *itemOptions) {
		o.expiration = exp
	}
}

// NewItem creates a new item with specified any options.
func NewItem[K comparable, V any](key K, val V, opts ...ItemOption) *Item[K, V] {
	o := new(itemOptions)
	for _, optFunc := range opts {
		optFunc(o)
	}
	now := nowFunc()
	return &Item[K, V]{
		Key:            key,
		Value:          val,
		ReferenceCount: 1,
		Expiration:     o.expiration,
		CreatedAt:      now,
		ReferencedAt:   now,
	}
}

// NumberCache is a in-memory cache which is able to store only Number constraint.
type NumberCache[K comparable, V Number] struct {
	Cache[K, V]
	// nmu is used to do lock in Increment/Decrement process.
	// Note that this must be here as a separate mutex because mu in Cache struct is Locked in GetItem,
	// and if we call mu.Lock in Increment/Decrement, it will cause deadlock.
	nmu sync.Mutex
}

// NewNumber creates a new cache for Number constraint.
func NewNumber[K comparable, V Number](baseCache Cache[K, V]) *NumberCache[K, V] {
	return &NumberCache[K, V]{
		Cache: baseCache,
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
