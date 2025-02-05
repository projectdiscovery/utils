package swissmap

import (
	"cmp"
	"reflect"
	"slices"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/cockroachdb/swiss"
)

// ComparableOrdered is an interface that extends [cmp.Ordered] with the
// comparable interface
type ComparableOrdered interface {
	cmp.Ordered
	comparable
}

// Map is a generic map implementation using swiss.Map with additional [Option]s
type Map[K ComparableOrdered, V any] struct {
	api        sonic.API
	data       *swiss.Map[K, V]
	mutex      sync.RWMutex
	threadSafe bool
	sorted     bool
	keys       []K
}

// New creates a new Map with the given options
func New[K ComparableOrdered, V any](options ...Option[K, V]) *Map[K, V] {
	m := &Map[K, V]{
		data: swiss.New[K, V](0),
		api:  getDefaultSonicConfig().Froze(),
	}

	for _, opt := range options {
		opt(m)
	}

	// TODO(dwisiswant0): Add check for comparable key type
	// if m.sorted {
	// 	var k K
	// 	if !reflect.TypeOf(k).Comparable() {
	// 		panic("key type must be comparable for sorted map")
	// 	}
	// }

	return m
}

// Clear removes all elements from the map
func (m *Map[K, V]) Clear() bool {
	m.lock()
	defer m.unlock()

	hadElements := m.data.Len() > 0

	m.data.Clear()
	m.keys = []K{}

	return hadElements
}

// Clone returns a new Map with a copy of the underlying data
func (m *Map[K, V]) Clone() *Map[K, V] {
	m.rLock()
	defer m.rUnlock()

	clone := New[K, V]()
	m.data.All(func(key K, value V) bool {
		clone.data.Put(key, value)

		return true
	})

	return clone
}

// Get retrieves a value from the map
func (m *Map[K, V]) Get(key K) (V, bool) {
	m.rLock()
	defer m.rUnlock()

	return m.data.Get(key)
}

// GetKeyWithValue retrieves the first key associated with the given value
func (m *Map[K, V]) GetKeyWithValue(value V) (K, bool) {
	m.rLock()
	defer m.rUnlock()

	var foundKey K
	var found bool

	m.data.All(func(key K, v V) bool {
		if reflect.DeepEqual(v, value) {
			foundKey = key
			found = true

			return false // stop iteration
		}

		return true
	})

	return foundKey, found
}

// GetKeys returns values for the given keys
func (m *Map[K, V]) GetKeys(keys ...K) []V {
	m.rLock()
	defer m.rUnlock()

	result := make([]V, 0, len(keys))
	for _, key := range keys {
		if val, ok := m.data.Get(key); ok {
			result = append(result, val)
		}
	}

	return result
}

// GetOrDefault returns the value for key or defaultValue if key is not found
func (m *Map[K, V]) GetOrDefault(key K, defaultValue V) V {
	m.rLock()
	defer m.rUnlock()

	if val, ok := m.data.Get(key); ok {
		return val
	}

	return defaultValue
}

// GetByIndex retrieves a value by its index
//
// The index is 0-based and must be less than the number of elements in the map
func (m *Map[K, V]) GetByIndex(idx int) (V, bool) {
	m.rLock()
	defer m.rUnlock()

	var value V

	// Return early if index out of range
	if idx < 0 || idx >= m.data.Len() {
		return value, false
	}

	return m.data.Get(m.keys[idx])
}

// Has checks if a key exists in the map
func (m *Map[K, V]) Has(key K) bool {
	m.rLock()
	defer m.rUnlock()

	_, ok := m.data.Get(key)

	return ok
}

// IsEmpty returns true if the map contains no elements
func (m *Map[K, V]) IsEmpty() bool {
	m.rLock()
	defer m.rUnlock()

	return m.data.Len() == 0
}

// Merge adds all key/value pairs from the input map
func (m *Map[K, V]) Merge(n map[K]V) {
	for k, v := range n {
		m.Set(k, v)
	}
}

// Set inserts or updates a key/value pair
func (m *Map[K, V]) Set(key K, value V) {
	m.lock()
	defer m.unlock()

	if !m.Has(key) {
		m.keys = append(m.keys, key)
		if m.sorted {
			// NOTE(dwisiswant0): It may cause a panic if the key is not comparable
			slices.SortStableFunc(m.keys, func(a, b K) int {
				return cmp.Compare(a, b)
			})
		}
	}

	m.data.Put(key, value)
}

// MarshalJSON marshals the map to JSON
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	m.rLock()
	defer m.rUnlock()

	target := make(map[K]V)

	m.data.All(func(key K, value V) bool {
		target[key] = value

		return true
	})

	return m.api.Marshal(target)
}

// UnmarshalJSON unmarshals the map from JSON
//
// The map is merged with the input data.
func (m *Map[K, V]) UnmarshalJSON(buf []byte) error {
	m.lock()
	defer m.unlock()

	target := make(map[K]V)

	if err := m.api.Unmarshal(buf, &target); err != nil {
		return err
	}

	m.Merge(target)

	return nil
}
