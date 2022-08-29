package cache

import (
	"context"
	"sync"
	"time"
)

// janitor for collecting expired items and cleaning them.
type janitor struct {
	ctx      context.Context
	interval time.Duration
	done     chan any
	once     sync.Once
}

// newJanitor creates a new janitor instance with the given interval variable
func newJanitor(ctx context.Context, interval time.Duration) *janitor {
	j := &janitor{
		ctx:      ctx,
		interval: interval,
		done:     make(chan any),
	}
	return j
}

// stop to stop the janitor
func (j *janitor) stop() {
	j.once.Do(func() { close(j.done) })
}

// run with the given cleanup callback function
func (j *janitor) run(cleanup func()) {
	go func() {
		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleanup()
			case <-j.done:
				cleanup() // last call
				return
			case <-j.ctx.Done():
				j.stop()
			}
		}
	}()
}
