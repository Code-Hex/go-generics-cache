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
