package cache

import (
	"sync"
)

// Cache is a common-cache interface.
type Cache[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, val V)
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
