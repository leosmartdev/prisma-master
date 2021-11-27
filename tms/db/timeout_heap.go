package db

import (
	"prisma/tms/log"

	"container/heap"
	"time"
)

// An object which can timeout
type TimeoutInfo interface {
	// Get the deadline for this object
	Deadline() time.Time
	// Get a unique identifier for this object
	ID() interface{}
}

/**
 * A heap which stores things which can timeout. Can be updated with updated
 * timeout times for each thing, can be queried for when the next timeout
 * occurs, and the next timeout can be pop'ed from the heap.
 */
type TimeoutHeap struct {
	idxs map[interface{}]int
	heap []TimeoutInfo
}

func NewTimeoutHeap() *TimeoutHeap {
	ret := &TimeoutHeap{
		idxs: make(map[interface{}]int),
		heap: make([]TimeoutInfo, 0, 128),
	}
	heap.Init(ret)
	return ret
}

// When/what is the next timeout
func (h *TimeoutHeap) Peek() (TimeoutInfo, bool) {
	if len(h.heap) == 0 {
		return nil, false
	}
	return h.heap[0], true
}

// Update or insert a timeout. Returns 'true' on update, 'false' on insert
func (h *TimeoutHeap) Upsert(info TimeoutInfo) bool {
	if idx, ok := h.idxs[info.ID()]; ok {
		if info.ID() != h.heap[idx].ID() {
			log.Fatal("Error looking up in heap!")
		}
		h.heap[idx] = info
		heap.Fix(h, idx)
		return true
	} else {
		heap.Push(h, info)
		return false
	}
}

// Is this object in the heap?
func (h *TimeoutHeap) Exists(info TimeoutInfo) bool {
	_, ok := h.idxs[info.ID()]
	return ok
}

// Needed for container/heap
func (h *TimeoutHeap) Len() int {
	return len(h.heap)
}

// Needed for container/heap
func (h *TimeoutHeap) Less(i, j int) bool {
	return h.heap[i].Deadline().Before(h.heap[j].Deadline())
}

// Needed for container/heap
func (h *TimeoutHeap) Swap(i, j int) {
	h.heap[i], h.heap[j] = h.heap[j], h.heap[i]
	h.idxs[h.heap[i].ID()] = i
	h.idxs[h.heap[j].ID()] = j
}

// Needed for container/heap
func (h *TimeoutHeap) Push(x interface{}) {
	n := len(h.heap)
	info := x.(TimeoutInfo)
	h.idxs[info.ID()] = n
	h.heap = append(h.heap, info)
}

// Needed for container/heap
func (h *TimeoutHeap) Pop() interface{} {
	old := h.heap
	n := len(old)
	info := old[n-1]
	delete(h.idxs, info.ID())
	h.heap = old[0 : n-1]
	return info
}
