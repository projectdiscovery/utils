package sliceutil

import "sync"

// SyncLock adds sync and lock capabilities to generic map
type SyncSlice[K comparable] struct {
	Slice []K
	mu    *sync.RWMutex
}

func NewSyncSlice[K comparable]() *SyncSlice[K] {
	return &SyncSlice[K]{mu: &sync.RWMutex{}}
}

func (ss *SyncSlice[K]) Append(items ...K) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.Slice = append(ss.Slice, items...)
}

func (ss *SyncSlice[K]) Empty() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.Slice = make([]K, 0)
}
