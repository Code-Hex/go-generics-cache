package cache

import (
	"context"
	"testing"
	"time"
)

func TestJanitor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	janitor := newJanitor(ctx, time.Millisecond)

	checkDone := make(chan struct{})
	janitor.done = checkDone

	calledClean := false
	janitor.run(func() { calledClean = true })

	// waiting for cleanup
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-checkDone:
	case <-time.After(time.Second):
		t.Fatalf("failed to call done channel")
	}

	if !calledClean {
		t.Fatal("failed to call clean callback in janitor")
	}
}
