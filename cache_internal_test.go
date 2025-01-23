package cache

import (
	"context"
	"testing"
	"time"
)

func TestDeletedCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc := NewContext[string, int](ctx)
	key := "key"
	nc.Set(key, 1, WithExpiration(-time.Second))

	_, ok := nc.cache.Get(key)
	if !ok {
		t.Fatal("want true")
	}

	nc.DeleteExpired()

	_, ok = nc.cache.Get(key)
	if ok {
		t.Fatal("want false")
	}
}

func TestDeleteExpired(t *testing.T) {
	now := time.Now()
	restore := func() {
		nowFunc = time.Now
	}

	t.Run("normal", func(t *testing.T) {
		defer restore()
		c := New[string, int]()

		c.Set("0", 0)
		c.Set("1", 10, WithExpiration(10*time.Millisecond))
		c.Set("2", 20, WithExpiration(20*time.Millisecond))
		c.Set("3", 30, WithExpiration(30*time.Millisecond))
		c.Set("4", 40, WithExpiration(40*time.Millisecond))
		c.Set("5", 50)

		maxItems := c.Len()

		expItems := 2

		for i := 0; i <= maxItems; i++ {
			nowFunc = func() time.Time {
				// Advance time to expire some items
				advanced := time.Duration(i * 10)
				return now.Add(advanced * time.Millisecond).Add(time.Millisecond)
			}

			c.DeleteExpired()

			got := c.Len()
			want := max(maxItems-i, expItems)
			if want != got {
				t.Errorf("want %d items but got %d", want, got)
			}
		}
	})

	t.Run("with remove", func(t *testing.T) {
		defer restore()
		c := New[string, int]()

		c.Set("0", 0)
		c.Set("1", 10, WithExpiration(10*time.Millisecond))
		c.Set("2", 20, WithExpiration(20*time.Millisecond))

		c.Delete("1")

		nowFunc = func() time.Time {
			return now.Add(30 * time.Millisecond).Add(time.Millisecond)
		}

		c.DeleteExpired()

		keys := c.Keys()
		want := 1
		if want != len(keys) {
			t.Errorf("want %d items but got %d", want, len(keys))
		}
	})

	t.Run("with update", func(t *testing.T) {
		defer restore()
		c := New[string, int]()

		c.Set("0", 0)
		c.Set("1", 10, WithExpiration(10*time.Millisecond))
		c.Set("2", 20, WithExpiration(20*time.Millisecond))
		c.Set("1", 30, WithExpiration(30*time.Millisecond)) // update

		maxItems := c.Len()

		nowFunc = func() time.Time {
			return now.Add(10 * time.Millisecond).Add(time.Millisecond)
		}

		c.DeleteExpired()

		got1 := c.Len()
		want1 := maxItems
		if want1 != got1 {
			t.Errorf("want1 %d items but got1 %d", want1, got1)
		}

		nowFunc = func() time.Time {
			return now.Add(30 * time.Millisecond).Add(time.Millisecond)
		}

		c.DeleteExpired()

		got2 := c.Len()
		want2 := 1
		if want2 != got2 {
			t.Errorf("want2 %d items but got2 %d", want2, got2)
		}
	})

	t.Run("issue #51", func(t *testing.T) {
		defer restore()
		c := New[string, int]()

		c.Set("1", 10, WithExpiration(10*time.Millisecond))
		c.Set("2", 20, WithExpiration(20*time.Millisecond))
		c.Set("1", 30, WithExpiration(100*time.Millisecond)) // expected do not expired key "1"

		nowFunc = func() time.Time {
			return now.Add(30 * time.Millisecond).Add(time.Millisecond)
		}

		c.DeleteExpired()

		got := c.Len()
		if want := 1; want != got {
			t.Errorf("want %d items but got %d", want, got)
		}
	})

	t.Run("issue #64", func(t *testing.T) {
		defer restore()
		c := New[string, int]()
		c.Set("1", 4, WithExpiration(0))  // These should not be expired
		c.Set("2", 5, WithExpiration(-1)) // These should not be expired
		c.Set("3", 6, WithExpiration(1*time.Hour))

		want := true
		_, ok := c.Get("1")
		if ok != want {
			t.Errorf("want %t but got %t", want, ok)
		}

		_, ok = c.Get("2")
		if ok != want {
			t.Errorf("want %t but got %t", want, ok)
		}
		_, ok = c.Get("3")
		if ok != want {
			t.Errorf("want %t but got %t", want, ok)
		}

	})
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
