package swissmap

import (
	"sync"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("Basic operations", func(t *testing.T) {
		m := New[string, int]()

		// Test Set and Get
		m.Set("one", 1)
		if val, ok := m.Get("one"); !ok || val != 1 {
			t.Errorf("expected Get(\"one\") = (1, true), got (%v, %v)", val, ok)
		}

		// Test Has
		if !m.Has("one") {
			t.Error("Has(\"one\") should return true")
		}

		// Test GetOrDefault
		if val := m.GetOrDefault("two", 2); val != 2 {
			t.Errorf("expected GetOrDefault(\"two\", 2) = 2, got %v", val)
		}

		// Test IsEmpty
		if m.IsEmpty() {
			t.Error("IsEmpty() should return false")
		}

		// Test Clear
		if !m.Clear() {
			t.Error("Clear() should return true for non-empty map")
		}
		if !m.IsEmpty() {
			t.Error("map should be empty after Clear()")
		}
	})

	t.Run("Concurrent operations", func(t *testing.T) {
		m := New[int, int]()
		var wg sync.WaitGroup
		n := 1000

		// Concurrent writers
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()
				m.Set(val, val*2)
			}(i)
		}

		// Concurrent readers
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(val int) {
				defer wg.Done()
				m.Has(val)
				m.Get(val)
			}(i)
		}

		wg.Wait()

		// Verify results
		count := 0
		for i := 0; i < n; i++ {
			if val, ok := m.Get(i); ok && val == i*2 {
				count++
			}
		}
		if count != n {
			t.Errorf("expected %d elements, got %d", n, count)
		}
	})

	t.Run("Clone and Merge", func(t *testing.T) {
		m := New[string, int]()
		m.Set("a", 1)
		m.Set("b", 2)

		// Test Clone
		clone := m.Clone()
		if val, ok := clone.Get("a"); !ok || val != 1 {
			t.Error("Clone did not copy values correctly")
		}

		// Test Merge
		other := map[string]int{"c": 3, "d": 4}
		m.Merge(other)
		if val, ok := m.Get("c"); !ok || val != 3 {
			t.Error("Merge did not add new values correctly")
		}
	})

	t.Run("GetKeyWithValue", func(t *testing.T) {
		m := New[string, int]()
		m.Set("a", 1)
		m.Set("b", 2)

		if key, ok := m.GetKeyWithValue(1); !ok || key != "a" {
			t.Errorf("GetKeyWithValue(1) = (%v, %v), want (\"a\", true)", key, ok)
		}

		if _, ok := m.GetKeyWithValue(3); ok {
			t.Error("GetKeyWithValue(3) should return false")
		}
	})

	t.Run("GetKeys", func(t *testing.T) {
		m := New[string, int]()
		m.Set("a", 1)
		m.Set("b", 2)
		m.Set("c", 3)

		values := m.GetKeys("a", "c", "missing")
		if len(values) != 2 || values[0] != 1 || values[1] != 3 {
			t.Errorf("GetKeys returned unexpected values: %v", values)
		}
	})
}
