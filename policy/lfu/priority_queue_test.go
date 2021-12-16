package lfu

import (
	"container/heap"
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
