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

	t.Run("keepTTL", func(t *testing.T) {

		t.Run("incr must keep ttl", func(t *testing.T) {
			defer restore()
			c := NewNumber[string, int]()

			c.Set("1", 10, WithExpiration(10*time.Millisecond))
			c.Increment("1", 20)
			nowFunc = func() time.Time {
				return now.Add(30 * time.Millisecond).Add(time.Millisecond)
			}
			c.DeleteExpired()
			if c.Len() != 0 {
				t.Fail()
			}
		})

		testCase := func(t *testing.T, wantN int, opts ...ItemOption) {
			defer restore()
			c := NewNumber[string, int]()

			c.Set("1", 10, WithExpiration(10*time.Millisecond))
			c.Set("1", 20, opts...)
			nowFunc = func() time.Time {
				return now.Add(300 * time.Millisecond).Add(time.Millisecond)
			}
			c.DeleteExpired()
			if c.Len() != wantN {
				t.Fail()
			}
		}
		t.Run("must forever when default set", func(t *testing.T) {
			testCase(t, 1)
		})
		t.Run("must expired when KeepTTL=true", func(t *testing.T) {
			testCase(t, 0, WithKeepTTL())
		})
		t.Run("must forever when KeepTTL=false", func(t *testing.T) {
			testCase(t, 1, WithKeepTTL(false))
		})
	})
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
