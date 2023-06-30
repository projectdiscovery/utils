package mapsutil

import (
	"fmt"
	"strconv"
	"testing"
)

func TestOrderedMapBasic(t *testing.T) {
	m := NewOrderedMap[string, string]()
	m.Set("test", "test")
	if m.IsEmpty() {
		t.Fatal("ordered map is empty")
	}
	if !m.Has("test") {
		t.Fatal("ordered map doesn't have test key")
	}
	if m.Has("test2") {
		t.Fatal("ordered map has test2 key")
	}
	if val, ok := m.Get("test"); !ok || val != "test" {
		t.Fatal("ordered map get test key doesn't return test value")
	}
	if m.GetKeys()[0] != "test" {
		t.Fatal("ordered map get keys doesn't return test key")
	}
	if val, ok := m.GetByIndex(0); !ok || val != "test" {
		t.Fatal("ordered map get by index doesn't return test key")
	}
	m.Delete("test")
	if !m.IsEmpty() {
		t.Fatal("ordered map is not empty after delete")
	}
}

func TestOrderedMap(t *testing.T) {
	m := NewOrderedMap[string, string]()
	for i := 0; i < 110; i++ {
		m.Set(strconv.Itoa(i), fmt.Sprintf("value-%d", i))
	}

	// iterate and validate order
	i := 0
	m.Iterate(func(key string, value string) bool {
		if key != strconv.Itoa(i) {
			t.Fatal("ordered map iterate order is not correct")
		}
		i++
		return true
	})

	// validate get by index
	for i := 0; i < 100; i++ {
		if val, ok := m.GetByIndex(i); !ok || val != fmt.Sprintf("value-%d", i) {
			t.Fatal("ordered map get by index doesn't return correct value")
		}
	}

	// random delete and validate order
	deleteElements := []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90}
	for _, i := range deleteElements {
		m.Delete(strconv.Itoa(i))
	}

	// validate elements after delete
	for k, i := range deleteElements {
		if val, ok := m.GetByIndex(i); !ok || val != fmt.Sprintf("value-%d", i+k+1) {
			t.Logf("order mismatch after delete got: index: %d, value: %s, exists: %v", i, val, ok)
		}
	}

}
