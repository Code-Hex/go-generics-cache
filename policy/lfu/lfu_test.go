package lfu_test

import (
	"testing"

	"github.com/Code-Hex/go-generics-cache/policy/lfu"
)

type tmp struct {
	i int
}

func (t *tmp) GetReferenceCount() int { return t.i }

func TestSet(t *testing.T) {
	// set capacity is 1
	cache := lfu.NewCache[string, int](lfu.WithCapacity(1))
	cache.Set("foo", 1)
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length: %d", got)
	}
	if got, ok := cache.Get("foo"); got != 1 || !ok {
		t.Fatalf("invalid value got %d, cachehit %v", got, ok)
	}

	// if over the cap
	cache.Set("bar", 2)
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length: %d", got)
	}
	bar, ok := cache.Get("bar")
	if bar != 2 || !ok {
		t.Fatalf("invalid value bar %d, cachehit %v", bar, ok)
	}

	// checks deleted oldest
	if _, ok := cache.Get("foo"); ok {
		t.Fatalf("invalid delete oldest value foo %v", ok)
	}

	// valid: if over the cap but same key
	cache.Set("bar", 100)
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length: %d", got)
	}
	bar, ok = cache.Get("bar")
	if bar != 100 || !ok {
		t.Fatalf("invalid replacing value bar %d, cachehit %v", bar, ok)
	}

	t.Run("with initilal reference count", func(t *testing.T) {
		cache := lfu.NewCache[string, *tmp](lfu.WithCapacity(2))
		cache.Set("foo", &tmp{i: 10}) // the highest reference count
		cache.Set("foo2", &tmp{i: 2}) // expected eviction
		if got := cache.Len(); got != 2 {
			t.Fatalf("invalid length: %d", got)
		}

		cache.Set("foo3", &tmp{i: 3})

		// checks deleted the lowest reference count
		if _, ok := cache.Get("foo2"); ok {
			t.Fatalf("invalid delete oldest value foo2 %v", ok)
		}
		if _, ok := cache.Get("foo"); !ok {
			t.Fatalf("invalid value foo is not found")
		}
	})
}

func TestDelete(t *testing.T) {
	cache := lfu.NewCache[string, int](lfu.WithCapacity(2))
	cache.Set("foo", 1)
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length: %d", got)
	}

	cache.Delete("foo2")
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length after deleted does not exist key: %d", got)
	}

	cache.Delete("foo")
	if got := cache.Len(); got != 0 {
		t.Fatalf("invalid length after deleted: %d", got)
	}
	if _, ok := cache.Get("foo"); ok {
		t.Fatalf("invalid get after deleted %v", ok)
	}
}

// check don't panic
func TestIssue33(t *testing.T) {
	cache := lfu.NewCache[string, int](lfu.WithCapacity(2))
	cache.Set("foo", 1)
	cache.Set("foo2", 2)
	cache.Set("foo3", 3)

	cache.Delete("foo")
	cache.Delete("foo2")
	cache.Delete("foo3")
}
