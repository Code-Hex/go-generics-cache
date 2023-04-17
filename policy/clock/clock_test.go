package clock_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/Code-Hex/go-generics-cache/policy/clock"
)

type tmp struct {
	i int
}

func (t *tmp) GetReferenceCount() int { return t.i }

func TestSet(t *testing.T) {
	// set capacity is 1
	cache := clock.NewCache[string, int](clock.WithCapacity(1))
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
		t.Fatalf("invalid eviction the oldest value for foo %v", ok)
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
		cache := clock.NewCache[string, *tmp](clock.WithCapacity(2))
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
	cache := clock.NewCache[string, int](clock.WithCapacity(1))
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

func TestKeys(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cache := clock.NewCache[string, int]()
		if len(cache.Keys()) != 0 {
			t.Errorf("want number of keys 0, but got %d", cache.Len())
		}

		cache.Set("foo", 1)
		cache.Set("bar", 2)
		cache.Set("baz", 3)
		cache.Set("bar", 4) // again
		cache.Set("foo", 5) // again

		got := strings.Join(cache.Keys(), ",")
		want := strings.Join([]string{
			"foo",
			"bar",
			"baz",
		}, ",")
		if got != want {
			t.Errorf("want %q, but got %q", want, got)
		}
		if len(cache.Keys()) != cache.Len() {
			t.Errorf("want number of keys %d, but got %d", len(cache.Keys()), cache.Len())
		}
	})

	t.Run("with deletion", func(t *testing.T) {
		cache := clock.NewCache[string, int](clock.WithCapacity(4))
		cache.Set("foo", 1)
		cache.Set("bar", 2)
		cache.Set("baz", 3)

		cache.Delete("bar") // delete in middle

		got := strings.Join(cache.Keys(), ",")
		want := strings.Join([]string{
			"foo",
			"baz",
		}, ",")
		if got != want {
			t.Errorf("want %q, but got %q", want, got)
		}
		if len(cache.Keys()) != cache.Len() {
			t.Errorf("want number of keys %d, but got %d", len(cache.Keys()), cache.Len())
		}

		cache.Set("hoge", 4)
		cache.Set("fuga", 5) // over the cap. so expected to set "bar" position.

		got2 := strings.Join(cache.Keys(), ",")
		want2 := strings.Join([]string{
			"foo",
			"fuga",
			"baz",
			"hoge",
		}, ",")
		if got2 != want2 {
			t.Errorf("want2 %q, but got2 %q", want2, got2)
		}
		if len(cache.Keys()) != cache.Len() {
			t.Errorf("want2 number of keys %d, but got2 %d", len(cache.Keys()), cache.Len())
		}
	})
}

func TestIssue29(t *testing.T) {
	cap := 3
	cache := clock.NewCache[string, int](clock.WithCapacity(cap))
	for i := 0; i < cap; i++ {
		cache.Set(strconv.Itoa(i), i)
	}
	cache.Set(strconv.Itoa(cap), cap)

	keys := cache.Keys()
	if got := len(keys); cap != got {
		t.Errorf("want number of keys %d, but got %d", cap, got)
	}
	wantKeys := "3,1,2"
	gotKeys := strings.Join(keys, ",")
	if wantKeys != gotKeys {
		t.Errorf("want keys %q, but got keys %q", wantKeys, gotKeys)
	}
}
