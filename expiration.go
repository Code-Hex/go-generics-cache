package cache

import (
	"container/heap"
	"time"
)

type expirationManager[K any] struct {
	queue expirationQueue[K]
}

func newExpirationManager[K any]() *expirationManager[K] {
	q := make(expirationQueue[K], 0)
	heap.Init(&q)
	return &expirationManager[K]{
		queue: q,
	}
}

func (m *expirationManager[K]) updateByIndex(index int, expiration time.Time) {
	if index >= 0 && index < m.len() {
		m.queue[index].expiration = expiration
		heap.Fix(&m.queue, index)
	}
}
func (m *expirationManager[K]) updateByItem(item *expirationKey[K], expiration time.Time) {
	m.updateByIndex(item.index, expiration)
}

func (m *expirationManager[K]) push(val K, expiration time.Time) *expirationKey[K] {
	item := &expirationKey[K]{
		Val:        val,
		expiration: expiration,
	}
	heap.Push(&m.queue, item)
	return item
}

func (m *expirationManager[K]) len() int {
	return m.queue.Len()
}

func (m *expirationManager[K]) pop() *expirationKey[K] {
	v := heap.Pop(&m.queue)
	return v.(*expirationKey[K])
}
func (m *expirationManager[K]) top() *expirationKey[K] {
	if m.len() == 0 {
		return nil
	}
	return m.queue[0]
}

func (m *expirationManager[K]) removeByIndex(index int) {
	heap.Remove(&m.queue, index)
}
func (m *expirationManager[K]) removeByItem(item *expirationKey[K]) {
	m.removeByIndex(item.index)
}

type expirationKey[K any] struct {
	Val        K
	expiration time.Time
	index      int
}

// expirationQueue implements heap.Interface and holds CacheItems.
type expirationQueue[K any] []*expirationKey[K]

var _ heap.Interface = (*expirationQueue[int])(nil)

func (pq expirationQueue[K]) Len() int { return len(pq) }

func (pq expirationQueue[K]) Less(i, j int) bool {
	// We want Pop to give us the least based on expiration time, not the greater
	return pq[i].expiration.Before(pq[j].expiration)
}

func (pq expirationQueue[K]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *expirationQueue[K]) Push(x interface{}) {
	n := len(*pq)
	item := x.(*expirationKey[K])
	item.index = n
	*pq = append(*pq, item)
}

func (pq *expirationQueue[K]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // For safety
	*pq = old[0 : n-1]
	return item
}
