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
	keys       []K
	mutex      sync.Mutex
	concurrent bool
	sorted     bool
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
	if m.lock() {
		defer m.unlock()
	}

	hadElements := m.data.Len() > 0
	m.data.Clear()

	// Reuse existing slice capacity
	if m.sorted {
		m.keys = m.keys[:0]
	}

	return hadElements
}

// Clone returns a new Map with a copy of the underlying data
func (m *Map[K, V]) Clone() *Map[K, V] {
	clone := New[K, V]()
	m.data.All(func(key K, value V) bool {
		clone.data.Put(key, value)

		return true
	})

	return clone
}

// Get retrieves a value from the map
func (m *Map[K, V]) Get(key K) (V, bool) {
	return m.data.Get(key)
}

// GetKeyWithValue retrieves the first key associated with the given value
func (m *Map[K, V]) GetKeyWithValue(value V) (K, bool) {
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
	if val, ok := m.data.Get(key); ok {
		return val
	}

	return defaultValue
}

// GetByIndex retrieves a value by its index
//
// The index is 0-based and must be less than the number of elements in the map
func (m *Map[K, V]) GetByIndex(idx int) (V, bool) {
	var value V
	var ok bool = false

	// Return early if index out of range
	if idx < 0 || idx >= m.data.Len() {
		return value, ok
	}

	if m.sorted {
		value, _ = m.data.Get(m.keys[idx])
		ok = true
	} else {
		i := 0
		m.data.All(func(key K, val V) bool {
			if i == idx {
				value = val
				return false
			}

			i++

			return true
		})

		ok = (i == idx)
	}

	return value, ok
}

// Has checks if a key exists in the map
func (m *Map[K, V]) Has(key K) bool {
	_, ok := m.data.Get(key)

	return ok
}

// IsEmpty returns true if the map contains no elements
func (m *Map[K, V]) IsEmpty() bool {
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
	if m.lock() {
		defer m.unlock()
	}

	m.data.Put(key, value)

	if m.sorted {
		if exists := m.Has(key); !exists {
			m.keys = append(m.keys, key)
			// NOTE(dwisiswant0): It may cause a panic if the key is not comparable
			slices.SortStableFunc(m.keys, func(a, b K) int {
				return cmp.Compare(a, b)
			})
		}
	}
}

// Iterate iterates over the [Map]
func (m *Map[K, V]) Iterate(fn func(key K, value V) bool) {
	if m.sorted {
		for _, key := range m.keys {
			value, ok := m.data.Get(key)
			if ok && !fn(key, value) {
				break
			}
		}
	} else {
		m.data.All(func(key K, value V) bool {
			return fn(key, value)
		})
	}
}

// MarshalJSON marshals the map to JSON
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	target := make(map[K]V, m.data.Len())

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
	if m.lock() {
		defer m.unlock()
	}

	target := make(map[K]V)

	if err := m.api.Unmarshal(buf, &target); err != nil {
		return err
	}

	m.Merge(target)

	return nil
}
