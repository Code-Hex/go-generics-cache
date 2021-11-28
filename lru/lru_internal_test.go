package lru

import (
	"testing"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

func TestContains(t *testing.T) {
	t.Run("without expiration", func(t *testing.T) {
		cache := NewCache[string, int]()
		cache.Set("foo", 1)
		cache.Set("bar", 2)
		cache.Set("baz", 3)
		for _, key := range []string{
			"foo",
			"bar",
			"baz",
		} {
			if !cache.Contains(key) {
				t.Errorf("not found: %s", key)
			}
		}
		if cache.Contains("not found") {
			t.Errorf("found")
		}
	})

	t.Run("with expiration", func(t *testing.T) {
		c := NewCache[string, int]()
		key := "foo"
		exp := time.Hour
		c.Set(key, 1, cache.WithExpiration(exp))
		// modify directly
		e, ok := c.items[key]
		if !ok {
			t.Fatal("unexpected not found key")
		}
		item := e.Value.(*cache.Item[string, int])
		item.CreatedAt = time.Now().Add(-2 * exp)

		if c.Contains(key) {
			t.Errorf("found")
		}
	})
}

func TestGet(t *testing.T) {
	c := NewCache[string, int]()
	key := "foo"
	exp := time.Hour
	c.Set(key, 1, cache.WithExpiration(exp))
	_, ok := c.Get(key)
	if !ok {
		t.Fatal("unexpected not found")
	}
	// modify directly
	e, ok := c.items[key]
	if !ok {
		t.Fatal("unexpected not found key")
	}
	item := e.Value.(*cache.Item[string, int])
	item.CreatedAt = time.Now().Add(-2 * exp)
	_, ok2 := c.Get(key)
	if ok2 {
		t.Fatal("unexpected found (expired)")
	}
}
