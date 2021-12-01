package lfu

import (
	"container/heap"

	cache "github.com/Code-Hex/go-generics-cache"
)

type entry[K comparable, V any] struct {
	index int
	item  *cache.Item[K, V]
}

func newEntry[K comparable, V any](key K, val V, opts ...cache.ItemOption) *entry[K, V] {
	return &entry[K, V]{
		index: 0,
		item:  cache.NewItem(key, val, opts...),
	}
}

type priorityQueue[K comparable, V any] []*entry[K, V]

func newPriorityQueue[K comparable, V any](cap int) *priorityQueue[K, V] {
	queue := make(priorityQueue[K, V], 0, cap)
	return &queue
}

// see example of priority queue: https://pkg.go.dev/container/heap
var _ heap.Interface = (*priorityQueue[interface{}, interface{}])(nil)

func (l priorityQueue[K, V]) Len() int { return len(l) }

func (l priorityQueue[K, V]) Less(i, j int) bool {
	if l[i].item.ReferenceCount == l[j].item.ReferenceCount {
		return l[i].item.ReferencedAt.Before(l[j].item.ReferencedAt)
	}
	return l[i].item.ReferenceCount < l[j].item.ReferenceCount
}

func (l priorityQueue[K, V]) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
	l[i].index = i
	l[i].index = j
}

func (l *priorityQueue[K, V]) Push(x interface{}) {
	entry := x.(*entry[K, V])
	entry.index = len(*l)
	*l = append(*l, entry)
}

func (l *priorityQueue[K, V]) Pop() interface{} {
	old := *l
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil   // avoid memory leak
	entry.index = -1 // for safety
	*l = old[0 : n-1]
	return entry
}

func (pq *priorityQueue[K, V]) update(e *entry[K, V], val V) {
	e.item.Value = val
	e.item.Referenced()
	heap.Fix(pq, e.index)
}