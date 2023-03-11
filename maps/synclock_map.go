package mapsutil

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrReadOnly = errors.New("read only mode")
)

// SyncLock adds sync and lock capabilities to generic map
type SyncLockMap[K, V comparable] struct {
	ReadOnly atomic.Bool
	mu       sync.RWMutex
	Map      Map[K, V]
}

// Lock the current map to read-only mode
func (s *SyncLockMap[K, V]) Lock() {
	s.ReadOnly.Store(true)
}

// Unlock the current map
func (s *SyncLockMap[K, V]) Unlock() {
	s.ReadOnly.Store(false)
}

// Set an item with syncronous access
func (s *SyncLockMap[K, V]) Set(k K, v V) error {
	if s.ReadOnly.Load() {
		return ErrReadOnly
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Map[k] = v

	return nil
}

// Get an item with syncronous access
func (s *SyncLockMap[K, V]) Get(k K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.Map[k]

	return v, ok
}

// Iterate with a callback function synchronously
func (s *SyncLockMap[K, V]) Iterate(f func(k K, v V) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k, v := range s.Map {
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}
