package lfu

import (
	"container/heap"
	"reflect"
	"testing"
	"time"
)

func TestPriorityQueue(t *testing.T) {
	// perl -MList::Util -e 'print join ",", List::Util::shuffle(1..10)'
	nums := []int{2, 1, 4, 5, 6, 9, 7, 10, 8, 3}
	queue := newPriorityQueue[int, int](len(nums))
	entries := make([]*entry[int, int], 0, len(nums))

	for _, v := range nums {
		entry := newEntry(v, v)
		entries = append(entries, entry)
		heap.Push(queue, entry)
	}

	if got := queue.Len(); len(nums) != got {
		t.Errorf("want %d, but got %d", len(nums), got)
	}

	// check the initial state
	for idx, entry := range *queue {
		if entry.index != idx {
			t.Errorf("want index %d, but got %d", entry.index, idx)
		}
		if entry.referenceCount != 1 {
			t.Errorf("want count 1")
		}
		if got := entry.val; nums[idx] != got {
			t.Errorf("want value %d but got %d", nums[idx], got)
		}
	}

	// updates len - 1 entries (updated all reference count and referenced_at)
	// so the lowest priority will be the last element.
	//
	// this loop creates
	// - Reference counters other than the last element are 2.
	// - The first element is the oldest referenced_at in reference counter is 2
	for i := 0; i < len(nums)-1; i++ {
		entry := entries[i]
		queue.update(entry, nums[i])
		time.Sleep(time.Millisecond)
	}

	// check the priority by reference counter
	wantValue := nums[len(nums)-1]
	got := heap.Pop(queue).(*entry[int, int])
	if got.index != -1 {
		t.Errorf("want index -1, but got %d", got.index)
	}
	if wantValue != got.val {
		t.Errorf("want the lowest priority value is %d, but got %d", wantValue, got.val)
	}
	if want, got := len(nums)-1, queue.Len(); want != got {
		t.Errorf("want %d, but got %d", want, got)
	}

	// check the priority by referenced_at
	wantValue2 := nums[0]
	got2 := heap.Pop(queue).(*entry[int, int])
	if got.index != -1 {
		t.Errorf("want index -1, but got %d", got.index)
	}
	if wantValue2 != got2.val {
		t.Errorf("want the lowest priority value is %d, but got %d", wantValue2, got2.val)
	}
	if want, got := len(nums)-2, queue.Len(); want != got {
		t.Errorf("want %d, but got %d", want, got)
	}
}

func Test_priorityQueue_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	type testCase[K comparable, V any] struct {
		name string
		q    *priorityQueue[K, V]
		args args
		want *priorityQueue[K, V]
	}
	tests := []testCase[string, int]{
		{
			name: "swap case",
			q: func() *priorityQueue[string, int] {
				q := newPriorityQueue[string, int](10)
				q.Push(&entry[string, int]{index: 0})
				q.Push(&entry[string, int]{index: 1})
				return q
			}(),
			args: args{i: 0, j: 1},
			want: func() *priorityQueue[string, int] {
				q := newPriorityQueue[string, int](10)
				q.Push(&entry[string, int]{index: 1})
				q.Push(&entry[string, int]{index: 0})
				return q
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.q.Swap(tt.args.i, tt.args.j)
			if !reflect.DeepEqual(tt.q, tt.want) {
				t.Errorf("want %v, got %v", tt.want, tt.q)
			}
		})
	}
}

func TestPriorityQueue_Pop(t *testing.T) {
	t.Run("Pop from empty queue", func(t *testing.T) {
		pq := newPriorityQueue[int, string](0)
		if elem := heap.Pop(pq); elem != nil {
			t.Errorf("Expected nil from empty queue, got %v", elem)
		}
	})

	t.Run("Pop from queue with single element", func(t *testing.T) {
		pq := newPriorityQueue[int, string](10)
		heap.Push(pq, newEntry(1, "one"))
		if pq.Len() != 1 {
			t.Fatalf("Expected queue length of 1, got %d", pq.Len())
		}
		elem := heap.Pop(pq).(*entry[int, string])
		if elem.key != 1 || elem.val != "one" {
			t.Errorf("Expected to pop element with key=1 and val='one', got key=%d and val='%s'", elem.key, elem.val)
		}
		if pq.Len() != 0 {
			t.Errorf("Expected empty queue after pop, got length %d", pq.Len())
		}
	})

	t.Run("Pop from queue with multiple elements", func(t *testing.T) {
		pq := newPriorityQueue[int, string](10)
		heap.Push(pq, newEntry(1, "one"))
		heap.Push(pq, newEntry(2, "two"))
		heap.Push(pq, newEntry(3, "three"))

		// Pop the first element
		elem := heap.Pop(pq).(*entry[int, string])
		if elem.key != 1 || elem.val != "one" {
			t.Errorf("Expected to pop element with key=1 and val='one', got key=%d and val='%s'", elem.key, elem.val)
		}

		// Pop the second element
		elem = heap.Pop(pq).(*entry[int, string])
		if elem.key != 2 || elem.val != "two" {
			t.Errorf("Expected to pop element with key=2 and val='two', got key=%d and val='%s'", elem.key, elem.val)
		}

		// Pop the third element
		elem = heap.Pop(pq).(*entry[int, string])
		if elem.key != 3 || elem.val != "three" {
			t.Errorf("Expected to pop element with key=3 and val='three', got key=%d and val='%s'", elem.key, elem.val)
		}

		if pq.Len() != 0 {
			t.Errorf("Expected empty queue after all pops, got length %d", pq.Len())
		}
	})
}
