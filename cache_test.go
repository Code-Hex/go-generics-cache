package cache_test

import (
	"math/rand"
	"sync"
	"testing"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/policy/clock"
	"github.com/Code-Hex/go-generics-cache/policy/fifo"
	"github.com/Code-Hex/go-generics-cache/policy/lfu"
	"github.com/Code-Hex/go-generics-cache/policy/lru"
	"github.com/Code-Hex/go-generics-cache/policy/mru"
)

func TestMultiThreadIncr(t *testing.T) {
	nc := cache.NewNumber[string, int]()
	nc.Set("counter", 0)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			_ = nc.Increment("counter", 1)
			wg.Done()
		}()
	}

	wg.Wait()

	if counter, _ := nc.Get("counter"); counter != 100 {
		t.Errorf("want %v but got %v", 100, counter)
	}
}

func TestMultiThreadDecr(t *testing.T) {
	nc := cache.NewNumber[string, int]()
	nc.Set("counter", 100)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			_ = nc.Decrement("counter", 1)
			wg.Done()
		}()
	}

	wg.Wait()

	if counter, _ := nc.Get("counter"); counter != 0 {
		t.Errorf("want %v but got %v", 0, counter)
	}
}

func TestMultiThread(t *testing.T) {
	cases := []struct {
		name   string
		policy cache.Option[int, int]
	}{
		{
			name:   "LRU",
			policy: cache.AsLRU[int, int](lru.WithCapacity(10)),
		},
		{
			name:   "MRU",
			policy: cache.AsMRU[int, int](mru.WithCapacity(10)),
		},
		{
			name:   "FIFO",
			policy: cache.AsFIFO[int, int](fifo.WithCapacity(10)),
		},
		{
			name:   "Clock",
			policy: cache.AsClock[int, int](clock.WithCapacity(10)),
		},
		{
			name:   "LFU",
			policy: cache.AsLFU[int, int](lfu.WithCapacity(10)),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := cache.New(tc.policy)
			var wg sync.WaitGroup
			for i := int64(0); i < 100; i++ {
				wg.Add(1)
				go func(i int64) {
					defer wg.Done()
					m := rand.New(rand.NewSource(i))
					for n := 0; n < 100; n++ {
						key := m.Intn(100000)
						c.Set(key, m.Intn(100000))
						c.Get(key)
					}
				}(i)
			}

			wg.Wait()
		})
	}
}
