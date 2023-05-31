package lfu

import (
	"container/heap"
	"time"

	"github.com/Code-Hex/go-generics-cache/policy/internal/policyutil"
)

type entry[K comparable, V any] struct {
	Index          int
	Key            K
	Val            V
	ReferenceCount int
	ReferencedAt   time.Time
}

func newEntry[K comparable, V any](key K, val V) *entry[K, V] {
	return &entry[K, V]{
		Index:          0,
		Key:            key,
		Val:            val,
		ReferenceCount: policyutil.GetReferenceCount(val),
		ReferencedAt:   time.Now(),
	}
}

func (e *entry[K, V]) referenced() {
	e.ReferenceCount++
	e.ReferencedAt = time.Now()
}

type priorityQueue[K comparable, V any] []*entry[K, V]

func newPriorityQueue[K comparable, V any](cap int) *priorityQueue[K, V] {
	queue := make(priorityQueue[K, V], 0, cap)
	return &queue
}

// see example of priority queue: https://pkg.go.dev/container/heap
var _ heap.Interface = (*priorityQueue[struct{}, interface{}])(nil)

func (q priorityQueue[K, V]) Len() int { return len(q) }

func (q priorityQueue[K, V]) Less(i, j int) bool {
	if q[i].ReferenceCount == q[j].ReferenceCount {
		return q[i].ReferencedAt.Before(q[j].ReferencedAt)
	}
	return q[i].ReferenceCount < q[j].ReferenceCount
}

func (q priorityQueue[K, V]) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].Index = i
	q[j].Index = j
}

func (q *priorityQueue[K, V]) Push(x interface{}) {
	entry := x.(*entry[K, V])
	entry.Index = len(*q)
	*q = append(*q, entry)
}

func (q *priorityQueue[K, V]) Pop() interface{} {
	old := *q
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil   // avoid memory leak
	entry.Index = -1 // for safety
	new := old[0 : n-1]
	for i := 0; i < len(new); i++ {
		new[i].Index = i
	}
	*q = new
	return entry
}

func (q *priorityQueue[K, V]) update(e *entry[K, V], val V) {
	e.Val = val
	e.referenced()
	heap.Fix(q, e.Index)
}
