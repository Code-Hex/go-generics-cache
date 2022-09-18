package cache

import (
	"context"
	"runtime"
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

func TestFinalizeCache(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skipf("Skipping on non-amd64 machine")
	}

	done := make(chan struct{})
	wait := make(chan struct{})
	go func() {
		c := New[string, int]()
		c.janitor.run(func() {
			close(done)
		})
		c = nil
		close(wait)
	}()
	<-wait

	runtime.GC()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("expected to call a function which is set as finalizer")
	}
}
