package mapsutil

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrReadOnly = errors.New("read only mode")
)

// SyncLock adds sync and lock capabilities to generic kv store callbacks
type SyncLock struct {
	ReadOnly        atomic.Bool
	mu              sync.RWMutex
	GetCallback     func(k interface{}) (interface{}, bool)
	SetCallback     func(k, v interface{}) error
	IterateCallback func(f func(k, v interface{}) error) error
}

func (s *SyncLock) Lock() {
	s.ReadOnly.Store(true)
}

func (m *SyncLock) Unlock() {
	m.ReadOnly.Store(false)
}

func (m *SyncLock) Set(k, v any) error {
	if m.ReadOnly.Load() {
		return ErrReadOnly
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.SetCallback(k, v)
}

func (m *SyncLock) Get(k any) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.GetCallback(k)
}

func (m *SyncLock) Iterate(f func(k, v any) error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.IterateCallback(f)
}
