package cache_test

import (
	"sync"
	"testing"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/simple"
)

func TestMultiThreadIncr(t *testing.T) {
	nc := cache.NewNumber[string, int](simple.NewCache[string, int]())
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
	nc := cache.NewNumber[string, int](simple.NewCache[string, int]())
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

func TestHasExpired(t *testing.T) {
	cases := []struct {
		name      string
		exp       time.Duration
		createdAt time.Time
		current   time.Time
		want      bool
	}{
		// expiration == createdAt + exp
		{
			name: "item expiration is zero",
			want: false,
		},
		{
			name:      "item expiration > current time",
			exp:       time.Hour * 24,
			createdAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			current:   time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			want:      false,
		},
		{
			name:      "item expiration < current time",
			exp:       time.Hour * 24,
			createdAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			current:   time.Date(2009, time.November, 12, 23, 0, 0, 0, time.UTC),
			want:      true,
		},
		{
			name:      "item expiration == current time",
			exp:       time.Second,
			createdAt: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			current:   time.Date(2009, time.November, 10, 23, 0, 1, 0, time.UTC),
			want:      false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reset := cache.SetNowFunc(tc.current)
			defer reset()

			it := &cache.Item[int, int]{
				Expiration: tc.exp,
				CreatedAt:  tc.createdAt,
			}
			if got := it.HasExpired(); tc.want != got {
				t.Fatalf("want %v, but got %v", tc.want, got)
			}
		})
	}
}
