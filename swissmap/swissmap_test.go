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
		m := New(WithThreadSafety[int, int]())
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

func TestMapJSON(t *testing.T) {
	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		m := New[string, int]()
		m.Set("one", 1)
		m.Set("two", 2)
		m.Set("three", 3)

		// Test marshaling
		data, err := m.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
		t.Logf("marshaled data: %s", data)

		// Test unmarshaling into new map
		newMap := New[string, int]()
		err = newMap.UnmarshalJSON(data)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		data2, err := newMap.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed for unmarshaled map: %v", err)
		}

		t.Logf("unmarshaled data: %s", data2)

		// Verify contents
		expected := map[string]int{"one": 1, "two": 2, "three": 3}
		for k, v := range expected {
			if val, ok := newMap.Get(k); !ok || val != v {
				t.Errorf("expected %s=%d, got %d (exists: %v)", k, v, val, ok)
			}
		}
	})

	t.Run("WithSortMapKeys", func(t *testing.T) {
		// t.SkipNow()

		m := New(WithSortMapKeys[string, int]())

		// Insert items in random order
		m.Set("zebra", 1)
		m.Set("alpha", 2)
		m.Set("beta", 3)

		data, err := m.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed with sorted keys: %v", err)
		}
		t.Logf("marshaled data with sorted keys: %s", data)

		// Test getting by index with sorted keys
		if v, ok := m.GetByIndex(0); !ok || v != 2 {
			t.Errorf("GetByIndex(0) = (%v, %v), want (2, true)", v, ok)
		}

		if v, ok := m.GetByIndex(1); !ok || v != 3 {
			t.Errorf("GetByIndex(1) = (%v, %v), want (3, true)", v, ok)
		}

		if v, ok := m.GetByIndex(2); !ok || v != 1 {
			t.Errorf("GetByIndex(2) = (%v, %v), want (1, true)", v, ok)
		}

		// Test out of bounds index
		if _, ok := m.GetByIndex(3); ok {
			t.Error("GetByIndex(3) should return false for out of bounds index")
		}

		// Test negative index
		if _, ok := m.GetByIndex(-1); ok {
			t.Error("GetByIndex(-1) should return false for negative index")
		}
	})

	t.Run("Empty map", func(t *testing.T) {
		m := New[string, string]()

		data, err := m.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed for empty map: %v", err)
		}

		newMap := New[string, string]()
		err = newMap.UnmarshalJSON(data)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed for empty map: %v", err)
		}

		if !newMap.IsEmpty() {
			t.Error("unmarshaled map should be empty")
		}
	})

	t.Run("Complex types", func(t *testing.T) {
		type Complex struct {
			Name string
			ID   int
		}
		m := New[string, Complex]()
		m.Set("item1", Complex{ID: 1, Name: "test1"})
		m.Set("item2", Complex{ID: 2, Name: "test2"})

		data, err := m.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed for complex types: %v", err)
		}

		newMap := New[string, Complex]()
		err = newMap.UnmarshalJSON(data)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed for complex types: %v", err)
		}

		if val, ok := newMap.Get("item1"); !ok || val.ID != 1 || val.Name != "test1" {
			t.Error("complex type was not correctly unmarshaled")
		}
	})
}
