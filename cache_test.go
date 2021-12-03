package cache_test

import (
	"sync"
	"testing"

	cache "github.com/Code-Hex/go-generics-cache"
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
