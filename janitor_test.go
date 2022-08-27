package cache

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestJanitor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	janitor := newJanitor(ctx, time.Millisecond)

	checkDone := make(chan struct{})
	janitor.done = checkDone

	calledClean := int64(0)
	janitor.run(func() { atomic.AddInt64(&calledClean, 1) })

	// waiting for cleanup
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-checkDone:
	case <-time.After(time.Second):
		t.Fatalf("failed to call done channel")
	}

	got := atomic.LoadInt64(&calledClean)
	if got <= 1 {
		t.Fatalf("failed to call clean callback in janitor: %d", got)
	}
}
