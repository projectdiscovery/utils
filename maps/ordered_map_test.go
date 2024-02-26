package mapsutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestOrderedMapMarshalUnmarshal(t *testing.T) {
	t.Run("TestSimpleStringToStringMapping", func(t *testing.T) {
		orderedMap1 := NewOrderedMap[string, string]()
		orderedMap1.Set("name", "John Doe")
		orderedMap1.Set("occupation", "Software Developer")

		marshaled1, err := json.Marshal(orderedMap1)
		if err != nil {
			t.Fatalf("Failed to marshal orderedMap1: %v", err)
		}

		unmarshaled1 := NewOrderedMap[string, string]()
		err = json.Unmarshal(marshaled1, &unmarshaled1)
		if err != nil {
			t.Fatalf("Failed to unmarshal orderedMap1: %v", err)
		}

		if !reflect.DeepEqual(orderedMap1, unmarshaled1) {
			t.Fatal("Unmarshaled map is not equal to the original map for orderedMap1")
		}
	})

	t.Run("TestIntegerToStructMapping", func(t *testing.T) {
		type Employee struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		orderedMap2 := NewOrderedMap[int, Employee]()
		orderedMap2.Set(1, Employee{ID: 1, Name: "Alice"})
		orderedMap2.Set(2, Employee{ID: 2, Name: "Bob"})

		marshaled2, err := json.Marshal(orderedMap2)
		if err != nil {
			t.Fatalf("Failed to marshal orderedMap2: %v", err)
		}

		unmarshaled2 := NewOrderedMap[int, Employee]()
		err = json.Unmarshal(marshaled2, &unmarshaled2)
		if err != nil {
			t.Fatalf("Failed to unmarshal orderedMap2: %v", err)
		}

		if !reflect.DeepEqual(orderedMap2, unmarshaled2) {
			t.Fatal("Unmarshaled map is not equal to the original map for orderedMap2")
		}
	})

	t.Run("TestStringToSliceOfStringsMapping", func(t *testing.T) {
		orderedMap3 := NewOrderedMap[string, []string]()
		orderedMap3.Set("fruits", []string{"apple", "banana", "cherry"})
		orderedMap3.Set("vegetables", []string{"tomato", "potato", "carrot"})

		marshaled3, err := json.Marshal(orderedMap3)
		if err != nil {
			t.Fatalf("Failed to marshal orderedMap3: %v", err)
		}

		unmarshaled3 := NewOrderedMap[string, []string]()
		err = json.Unmarshal(marshaled3, &unmarshaled3)
		if err != nil {
			t.Fatalf("Failed to unmarshal orderedMap3: %v", err)
		}

		if !reflect.DeepEqual(orderedMap3, unmarshaled3) {
			t.Fatal("Unmarshaled map is not equal to the original map for orderedMap3")
		}
	})
}

func TestOrderedMapDeleteWhileIterating(t *testing.T) {
	om := NewOrderedMap[string, string]()
	om.Set("key1", "value1")
	om.Set("key2", "value2")
	om.Set("key3", "value3")

	ignoreKey := "key1"

	got := []string{}

	om.Iterate(func(key string, value string) bool {
		got = append(got, key)
		if key == ignoreKey {
			om.Delete(key)
		}
		return true
	})

	require.ElementsMatchf(t, []string{"key1", "key2", "key3"}, got, "inconsistent iteration order")
}
