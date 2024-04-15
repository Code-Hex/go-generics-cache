package cache

import (
	"context"
	"github.com/Code-Hex/go-generics-cache/policy/clock"
	"github.com/Code-Hex/go-generics-cache/policy/fifo"
	"github.com/Code-Hex/go-generics-cache/policy/lfu"
	"github.com/Code-Hex/go-generics-cache/policy/lru"
	"github.com/Code-Hex/go-generics-cache/policy/mru"
	"math/rand"
	"runtime"
	"runtime/debug"
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
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func compactGcBenchmark[K comparable, V any](b *testing.B, f func(cache *Cache[K, V]), newOpts ...Option[K, V]) {
	b.StopTimer()
	time.Sleep(0)
	runtime.GC()
	debug.FreeOSMemory()

	ctx, cancel := context.WithCancel(context.Background())
	cs := NewContext[K, V](ctx, newOpts...)
	b.StartTimer()
	f(cs)
	b.StopTimer()

	cancel()
	cs = nil
	f = nil
	time.Sleep(0)
	runtime.GC()
	debug.FreeOSMemory()
	b.StartTimer()
}

const (
	_ = iota
	caseTypeLRU
	caseTypeLFU
	caseTypeMRU
	caseTypeClock
	caseTypeFIFO
)

func allCacheBenchmark[K comparable, V any](b *testing.B,
	testCase func(b *testing.B, c *Cache[K, V]),
	withCapCount func(b *testing.B, tp int) int) {
	b.Run("simple", func(b *testing.B) {
		compactGcBenchmark(b, func(c *Cache[K, V]) {
			testCase(b, c)
		})
	})
	b.Run("LRU", func(b *testing.B) {
		cn := withCapCount(b, caseTypeLRU)
		compactGcBenchmark(b, func(c *Cache[K, V]) {
			testCase(b, c)
		}, AsLRU[K, V](lru.WithCapacity(cn)))
	})
	b.Run("LFU", func(b *testing.B) {
		cn := withCapCount(b, caseTypeLFU)
		compactGcBenchmark(b, func(c *Cache[K, V]) {
			testCase(b, c)
		}, AsLFU[K, V](lfu.WithCapacity(cn)))
	})
	b.Run("MRU", func(b *testing.B) {
		cn := withCapCount(b, caseTypeMRU)
		compactGcBenchmark(b, func(c *Cache[K, V]) {
			testCase(b, c)
		}, AsMRU[K, V](mru.WithCapacity(cn)))
	})
	b.Run("Clock", func(b *testing.B) {
		cn := withCapCount(b, caseTypeClock)
		compactGcBenchmark(b, func(c *Cache[K, V]) {
			testCase(b, c)
		}, AsClock[K, V](clock.WithCapacity(cn)))
	})
	b.Run("FIFO", func(b *testing.B) {
		cn := withCapCount(b, caseTypeFIFO)
		compactGcBenchmark(b, func(c *Cache[K, V]) {
			testCase(b, c)
		}, AsFIFO[K, V](fifo.WithCapacity(cn)))
	})
}

func compactGcTest[K comparable, V any](t *testing.T, f func(cache *Cache[K, V]), newOpts ...Option[K, V]) {
	time.Sleep(0)
	runtime.GC()
	debug.FreeOSMemory()

	ctx, cancel := context.WithCancel(context.Background())
	cs := NewContext[K, V](ctx, newOpts...)
	f(cs)

	cancel()
	cs = nil
	f = nil
	time.Sleep(0)
	runtime.GC()
	debug.FreeOSMemory()
}

func allCacheTest[K comparable, V any](t *testing.T,
	testCase func(t *testing.T, c *Cache[K, V]),
	withCapCount func(t *testing.T, tp int) int) {
	t.Run("simple", func(t *testing.T) {
		compactGcTest(t, func(c *Cache[K, V]) {
			testCase(t, c)
		})
	})
	t.Run("LRU", func(t *testing.T) {
		cn := withCapCount(t, caseTypeLRU)
		compactGcTest(t, func(c *Cache[K, V]) {
			testCase(t, c)
		}, AsLRU[K, V](lru.WithCapacity(cn)))
	})
	t.Run("LFU", func(t *testing.T) {
		cn := withCapCount(t, caseTypeLFU)
		compactGcTest(t, func(c *Cache[K, V]) {
			testCase(t, c)
		}, AsLFU[K, V](lfu.WithCapacity(cn)))
	})
	t.Run("MRU", func(t *testing.T) {
		cn := withCapCount(t, caseTypeMRU)
		compactGcTest(t, func(c *Cache[K, V]) {
			testCase(t, c)
		}, AsMRU[K, V](mru.WithCapacity(cn)))
	})
	t.Run("Clock", func(t *testing.T) {
		cn := withCapCount(t, caseTypeClock)
		compactGcTest(t, func(c *Cache[K, V]) {
			testCase(t, c)
		}, AsClock[K, V](clock.WithCapacity(cn)))
	})
	t.Run("FIFO", func(t *testing.T) {
		cn := withCapCount(t, caseTypeFIFO)
		compactGcTest(t, func(c *Cache[K, V]) {
			testCase(t, c)
		}, AsFIFO[K, V](fifo.WithCapacity(cn)))
	})
}

// var _testRandGen = rand.New(rand.NewSource(time.Now().Unix()))
var _testRandGen = rand.New(rand.NewSource(12345678))

func BenchmarkSet(b *testing.B) {
	const maxKeySize = 2000000 // -benchtime=10s use mem:~300-700mb
	allCacheBenchmark(b, func(b *testing.B, c *Cache[int, int]) {
		for i := 0; i < b.N; i++ {
			c.Set(_testRandGen.Int()%maxKeySize, i)
		}
	}, func(b *testing.B, tp int) int {
		return min(maxKeySize, b.N)
	})
}

func BenchmarkGetExist(b *testing.B) {
	const maxKeySize = 2000000 // -benchtime=10s use mem:~300-700mb
	allCacheBenchmark(b, func(b *testing.B, c *Cache[int, int]) {
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			c.Set(_testRandGen.Int()%maxKeySize, i)
		}
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			c.Get(_testRandGen.Int() % maxKeySize)
		}
	}, func(b *testing.B, tp int) int {
		return min(maxKeySize, b.N)
	})
}

func BenchmarkCacheSetTTL(b *testing.B) {
	const maxKeySize = 2000000 // -benchtime=10s use mem:~300-700mb
	allCacheBenchmark(b, func(b *testing.B, c *Cache[int, int]) {
		for i := 0; i < b.N; i++ {
			c.Set(_testRandGen.Int()%maxKeySize, i, WithExpiration(time.Hour))
		}
	}, func(b *testing.B, tp int) int {
		return min(maxKeySize, b.N)
	})
}

func TestTTLClear(t *testing.T) {
	const maxKeySize = 30000
	seed := _testRandGen.Int63()
	restore := func() {
		nowFunc = time.Now
	}
	allCacheTest(t, func(t *testing.T, c *Cache[int, int]) {
		_testRandGen.Seed(seed)
		nt := time.Now()
		nowFunc = func() time.Time { return nt }
		defer restore()
		for i := 0; i < maxKeySize; i++ {
			c.Set(_testRandGen.Int()%maxKeySize, i, WithExpiration(time.Second*time.Duration(_testRandGen.Int()%4000)))
		}
		nowFunc = func() time.Time { return nt.Add(time.Second * 2000) }
		take := time.Now()
		len1 := c.Len()
		c.DeleteExpired()
		n := len1 - c.Len()
		t.Log("clear expired", n, time.Now().Sub(take))

	}, func(b *testing.T, tp int) int {
		return maxKeySize
	})

}
