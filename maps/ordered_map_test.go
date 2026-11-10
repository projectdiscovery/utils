package mapsutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
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

// orderedMapKeys returns a non-nil copy of the keys so that tests can compare
// empty maps with require.Equal.
func orderedMapKeys(omap *OrderedMap[string, int]) []string {
	return append([]string{}, omap.GetKeys()...)
}

// requireOrderedMapConsistent checks that every read path agrees on the same
// live keys and order, including while Iterate is running.
func requireOrderedMapConsistent(t *testing.T, omap *OrderedMap[string, int]) {
	t.Helper()

	keys := orderedMapKeys(omap)
	require.Len(t, keys, omap.Len(), "GetKeys and Len disagree: %q", keys)
	require.Equal(t, len(keys) == 0, omap.IsEmpty(), "IsEmpty disagrees with GetKeys: %q", keys)

	seen := map[string]struct{}{}

	for idx, key := range keys {
		require.True(t, omap.Has(key), "GetKeys returned absent key %q (keys %q)", key, keys)
		_, dup := seen[key]
		require.False(t, dup, "GetKeys returned duplicate key %q (keys %q)", key, keys)

		seen[key] = struct{}{}

		want, _ := omap.Get(key)
		got, ok := omap.GetByIndex(idx)
		require.True(t, ok, "GetByIndex(%d) missing (keys %q)", idx, keys)
		require.Equal(t, want, got, "GetByIndex(%d) returned the wrong value (keys %q)", idx, keys)
	}

	_, ok := omap.GetByIndex(len(keys))
	require.False(t, ok, "GetByIndex past the last key succeeded (keys %q)", keys)

	clone := omap.Clone()
	require.Equal(t, keys, orderedMapKeys(&clone), "Clone keys differ")
	require.Equal(t, len(keys), clone.Len(), "Clone length differs")

	data, err := json.Marshal(omap)
	require.NoError(t, err)

	decoded := NewOrderedMap[string, int]()
	require.NoError(t, json.Unmarshal(data, &decoded), "marshaled form %s", data)
	require.Equal(t, keys, orderedMapKeys(&decoded), "marshaled form %s", data)
}

type orderedMapMutationCase struct {
	name        string
	initial     []string
	visit       func(omap *OrderedMap[string, int], key string) bool
	wantVisited []string
	wantKeys    []string
}

// onKey returns a visit function that runs act when the iteration reaches
// target and always continues.
func onKey(target string, act func(omap *OrderedMap[string, int])) func(*OrderedMap[string, int], string) bool {
	return func(omap *OrderedMap[string, int], key string) bool {
		if key == target {
			act(omap)
		}

		return true
	}
}

func orderedMapDeleteCases() []orderedMapMutationCase {
	return []orderedMapMutationCase{
		{
			name:    "delete unvisited key",
			initial: []string{"a", "b", "c"},
			visit: onKey("a", func(omap *OrderedMap[string, int]) {
				omap.Delete("b")
			}),
			wantVisited: []string{"a", "c"},
			wantKeys:    []string{"a", "c"},
		},
		{
			name:    "delete already visited key",
			initial: []string{"a", "b", "c"},
			visit: onKey("c", func(omap *OrderedMap[string, int]) {
				omap.Delete("a")
			}),
			wantVisited: []string{"a", "b", "c"},
			wantKeys:    []string{"b", "c"},
		},
		{
			name:    "delete current key",
			initial: []string{"a", "b", "c"},
			visit: onKey("b", func(omap *OrderedMap[string, int]) {
				omap.Delete("b")
			}),
			wantVisited: []string{"a", "b", "c"},
			wantKeys:    []string{"a", "c"},
		},
		{
			name:    "delete every key",
			initial: []string{"a", "b", "c"},
			visit: func(omap *OrderedMap[string, int], key string) bool {
				omap.Delete(key)

				return true
			},
			wantVisited: []string{"a", "b", "c"},
			wantKeys:    []string{},
		},
	}
}

func orderedMapSetCases() []orderedMapMutationCase {
	return []orderedMapMutationCase{
		{
			name:    "delete absent key",
			initial: []string{"a", "b"},
			visit: func(omap *OrderedMap[string, int], _ string) bool {
				omap.Delete("missing")

				return true
			},
			wantVisited: []string{"a", "b"},
			wantKeys:    []string{"a", "b"},
		},
		{
			name:    "delete same key twice",
			initial: []string{"a", "b", "c"},
			visit: onKey("a", func(omap *OrderedMap[string, int]) {
				omap.Delete("b")
				omap.Delete("b")
			}),
			wantVisited: []string{"a", "c"},
			wantKeys:    []string{"a", "c"},
		},
		{
			name:    "set new key is not visited",
			initial: []string{"a", "b"},
			visit: onKey("a", func(omap *OrderedMap[string, int]) {
				omap.Set("z", 26)
			}),
			wantVisited: []string{"a", "b"},
			wantKeys:    []string{"a", "b", "z"},
		},
		{
			name:    "set existing key keeps its position",
			initial: []string{"a", "b", "c"},
			visit: onKey("a", func(omap *OrderedMap[string, int]) {
				omap.Set("b", 20)
			}),
			wantVisited: []string{"a", "b", "c"},
			wantKeys:    []string{"a", "b", "c"},
		},
	}
}

func orderedMapReAddCases() []orderedMapMutationCase {
	return []orderedMapMutationCase{
		{
			name:    "delete then set unvisited key moves it to the end",
			initial: []string{"a", "b", "c"},
			visit: onKey("a", func(omap *OrderedMap[string, int]) {
				omap.Delete("b")
				omap.Set("b", 20)
			}),
			wantVisited: []string{"a", "b", "c"},
			wantKeys:    []string{"a", "c", "b"},
		},
		{
			name:    "delete then set visited key moves it to the end",
			initial: []string{"a", "b", "c"},
			visit: onKey("c", func(omap *OrderedMap[string, int]) {
				omap.Delete("a")
				omap.Set("a", 10)
			}),
			wantVisited: []string{"a", "b", "c"},
			wantKeys:    []string{"b", "c", "a"},
		},
		{
			name:    "rename keys",
			initial: []string{"a", "b", "c"},
			visit: func(omap *OrderedMap[string, int], key string) bool {
				value, _ := omap.Get(key)
				omap.Delete(key)
				omap.Set(key+"2", value)

				return true
			},
			wantVisited: []string{"a", "b", "c"},
			wantKeys:    []string{"a2", "b2", "c2"},
		},
		{
			// five Sets leave spare capacity in the keys slice that a running
			// Iterate shares, so appends must not touch the entries it reads
			name:    "interleaved sets and deletes with spare capacity",
			initial: []string{"a", "b", "c", "d", "e"},
			visit: onKey("a", func(omap *OrderedMap[string, int]) {
				omap.Set("f", 6)
				omap.Delete("b")
				omap.Set("g", 7)
				omap.Delete("c")
			}),
			wantVisited: []string{"a", "d", "e"},
			wantKeys:    []string{"a", "d", "e", "f", "g"},
		},
	}
}

func orderedMapEarlyStopAndNestedCases() []orderedMapMutationCase {
	return []orderedMapMutationCase{
		{
			name:    "early stop with pending delete",
			initial: []string{"a", "b", "c"},
			visit: func(omap *OrderedMap[string, int], _ string) bool {
				omap.Delete("b")

				return false
			},
			wantVisited: []string{"a"},
			wantKeys:    []string{"a", "c"},
		},
		{
			name:    "nested iterate without changes, then outer delete",
			initial: []string{"a", "b", "c", "d"},
			visit: onKey("a", func(omap *OrderedMap[string, int]) {
				omap.Iterate(func(string, int) bool { return true })
				omap.Delete("b")
			}),
			wantVisited: []string{"a", "c", "d"},
			wantKeys:    []string{"a", "c", "d"},
		},
		{
			name:    "nested iterate deletes, then outer deletes",
			initial: []string{"a", "b", "c", "d"},
			visit: onKey("a", func(omap *OrderedMap[string, int]) {
				omap.Iterate(func(string, int) bool {
					omap.Delete("c")

					return true
				})
				omap.Delete("b")
			}),
			wantVisited: []string{"a", "d"},
			wantKeys:    []string{"a", "d"},
		},
		{
			name:    "nested iterate stops early with pending delete",
			initial: []string{"a", "b", "c"},
			visit: onKey("a", func(omap *OrderedMap[string, int]) {
				omap.Iterate(func(string, int) bool {
					omap.Delete("b")

					return false
				})
			}),
			wantVisited: []string{"a", "c"},
			wantKeys:    []string{"a", "c"},
		},
	}
}

func orderedMapMutationCases() []orderedMapMutationCase {
	return slices.Concat(
		orderedMapDeleteCases(),
		orderedMapSetCases(),
		orderedMapReAddCases(),
		orderedMapEarlyStopAndNestedCases(),
	)
}

func TestOrderedMapMutateWhileIterating(t *testing.T) {
	t.Parallel()

	for _, tcase := range orderedMapMutationCases() {
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()

			omap := NewOrderedMap[string, int]()
			for i, key := range tcase.initial {
				omap.Set(key, i+1)
			}

			visited := []string{}

			omap.Iterate(func(key string, _ int) bool {
				visited = append(visited, key)
				cont := tcase.visit(&omap, key)

				requireOrderedMapConsistent(t, &omap)

				return cont
			})

			require.Equal(t, tcase.wantVisited, visited, "visited keys")
			require.Equal(t, tcase.wantKeys, orderedMapKeys(&omap), "keys after Iterate")
			require.Equal(t, len(tcase.wantKeys), omap.Len(), "Len after Iterate")
			requireOrderedMapConsistent(t, &omap)

			// The map must be back in its non-iterating mode: an immediate
			// Set and Delete take effect on the very next read.
			omap.Set("tail", 99)
			require.Equal(t, append(append([]string{}, tcase.wantKeys...), "tail"), orderedMapKeys(&omap))
			omap.Delete("tail")
			require.Equal(t, tcase.wantKeys, orderedMapKeys(&omap))
			requireOrderedMapConsistent(t, &omap)
		})
	}
}

// MarshalJSON has a value receiver, so it works on a copy that shares the
// pending deletions with the original.
func TestOrderedMapMarshalWhileIteratingKeepsPendingDeletes(t *testing.T) {
	t.Parallel()

	omap := NewOrderedMap[string, int]()
	omap.Set("a", 1)
	omap.Set("b", 2)
	omap.Set("c", 3)

	omap.Iterate(func(key string, _ int) bool {
		if key != "a" {
			return true
		}

		omap.Delete("b")
		omap.Set("b", 20)
		omap.Delete("c")

		data, err := json.Marshal(&omap)
		require.NoError(t, err)
		require.Equal(t, `{"a":1,"b":20}`, string(data))

		require.Equal(t, []string{"a", "b"}, orderedMapKeys(&omap))

		return true
	})
	require.Equal(t, []string{"a", "b"}, orderedMapKeys(&omap))
}

func TestOrderedMapIteratePanicRestoresState(t *testing.T) {
	t.Parallel()

	omap := NewOrderedMap[string, int]()
	omap.Set("a", 1)
	omap.Set("b", 2)
	omap.Set("c", 3)

	require.PanicsWithValue(t, "boom", func() {
		omap.Iterate(func(key string, _ int) bool {
			if key == "a" {
				omap.Delete("b")
				panic("boom")
			}

			return true
		})
	})

	require.Equal(t, []string{"a", "c"}, orderedMapKeys(&omap))
	omap.Delete("a")
	require.Equal(t, []string{"c"}, orderedMapKeys(&omap))
	requireOrderedMapConsistent(t, &omap)
}

func BenchmarkOrderedMapDelete(b *testing.B) {
	tests := []struct {
		name string
		run  func(omap *OrderedMap[int, int], size int)
	}{
		{"build only", func(*OrderedMap[int, int], int) {}},
		{"get each", func(omap *OrderedMap[int, int], size int) {
			for key := range size {
				omap.Get(key)
			}
		}},
		{"len each", func(omap *OrderedMap[int, int], size int) {
			for range size {
				omap.Len()
			}
		}},
		{"iterate", func(omap *OrderedMap[int, int], _ int) {
			omap.Iterate(func(int, int) bool { return true })
		}},
		{"delete each in iterate", func(omap *OrderedMap[int, int], _ int) {
			omap.Iterate(func(key, _ int) bool {
				omap.Delete(key)

				return true
			})
		}},
		{"rename each in iterate", func(omap *OrderedMap[int, int], _ int) {
			omap.Iterate(func(key, value int) bool {
				omap.Delete(key)
				omap.Set(-key-1, value)

				return true
			})
		}},
		{"delete each outside iterate", func(omap *OrderedMap[int, int], size int) {
			for key := range size {
				omap.Delete(key)
			}
		}},
	}

	for _, size := range []int{10, 100, 1000, 10000} {
		for _, tcase := range tests {
			b.Run(fmt.Sprintf("size=%d/%s", size, tcase.name), func(b *testing.B) {
				b.ReportAllocs()

				for range b.N {
					omap := NewOrderedMap[int, int]()
					for key := range size {
						omap.Set(key, key)
					}

					tcase.run(&omap, size)
				}
			})
		}
	}
}

type orderedMapNestedCase struct {
	name        string
	outer       func(omap *OrderedMap[string, int], inner *[]string)
	wantInner   []string
	wantOuter   []string
	wantKeysEnd []string
}

func orderedMapNestedCases() []orderedMapNestedCase {
	return []orderedMapNestedCase{
		{
			name: "outer deletes and sets an unvisited key before the inner iterate",
			outer: func(omap *OrderedMap[string, int], inner *[]string) {
				omap.Delete("b")
				omap.Set("b", 20)
				omap.Iterate(func(key string, _ int) bool {
					*inner = append(*inner, key)

					return true
				})
			},
			wantInner:   []string{"a", "c", "b"},
			wantOuter:   []string{"a", "b", "c"},
			wantKeysEnd: []string{"a", "c", "b"},
		},
		{
			name: "inner callback sets a key the outer callback deleted",
			outer: func(omap *OrderedMap[string, int], inner *[]string) {
				omap.Delete("b")
				omap.Iterate(func(key string, _ int) bool {
					*inner = append(*inner, key)
					if key == "a" {
						omap.Set("b", 20)
					}

					return true
				})
			},
			wantInner:   []string{"a", "c"},
			wantOuter:   []string{"a", "b", "c"},
			wantKeysEnd: []string{"a", "c", "b"},
		},
	}
}

func TestOrderedMapNestedIterateSeesLiveKeys(t *testing.T) {
	t.Parallel()

	for _, tcase := range orderedMapNestedCases() {
		t.Run(tcase.name, func(t *testing.T) {
			t.Parallel()

			omap := NewOrderedMap[string, int]()
			omap.Set("a", 1)
			omap.Set("b", 2)
			omap.Set("c", 3)

			inner, outer := []string{}, []string{}

			omap.Iterate(func(key string, _ int) bool {
				outer = append(outer, key)
				if key == "a" {
					tcase.outer(&omap, &inner)
				}

				return true
			})

			require.Equal(t, tcase.wantInner, inner, "inner visited keys")
			require.Equal(t, tcase.wantOuter, outer, "outer visited keys")
			require.Equal(t, tcase.wantKeysEnd, orderedMapKeys(&omap), "keys after Iterate")
			requireOrderedMapConsistent(t, &omap)
		})
	}
}

func TestOrderedMapUnmarshalWhileIterating(t *testing.T) {
	t.Parallel()

	omap := NewOrderedMap[string, int]()
	omap.Set("a", 1)
	omap.Set("b", 2)

	omap.Iterate(func(key string, _ int) bool {
		if key == "a" {
			omap.Delete("b")
			require.NoError(t, json.Unmarshal([]byte(`{"b":5,"x":1}`), &omap))
		}

		return true
	})

	require.Equal(t, []string{"b", "x"}, orderedMapKeys(&omap))
	require.Equal(t, 2, omap.Len())
	require.True(t, omap.Has("b"))
	require.True(t, omap.Has("x"))

	data, err := json.Marshal(&omap)
	require.NoError(t, err)
	require.Equal(t, `{"b":5,"x":1}`, string(data))
}

// UnmarshalJSON keeps repeated keys of a JSON object in keys, which is outside
// what OrderedMap guarantees. These tests pin the behavior of the read and
// Delete paths for such a map.
func TestOrderedMapDuplicateJSONKeys(t *testing.T) {
	t.Parallel()

	omap := NewOrderedMap[string, int]()
	require.NoError(t, json.Unmarshal([]byte(`{"a":1,"a":2,"b":3}`), &omap))

	require.Equal(t, []string{"a", "a", "b"}, orderedMapKeys(&omap))
	require.Equal(t, 3, omap.Len())

	for idx := range omap.Len() {
		_, ok := omap.GetByIndex(idx)
		require.True(t, ok, "GetByIndex(%d) of Len() %d", idx, omap.Len())
	}

	// each Delete removes one leftover entry, as long as one remains
	omap.Delete("a")
	require.Equal(t, []string{"a", "b"}, orderedMapKeys(&omap))
	require.Equal(t, 2, omap.Len())

	omap.Delete("a")
	require.Equal(t, []string{"b"}, orderedMapKeys(&omap))
	require.Equal(t, 1, omap.Len())
	require.False(t, omap.IsEmpty())

	omap.Delete("b")
	require.True(t, omap.IsEmpty())
}

// orderedMapAPI is the surface the differential test drives on both the
// OrderedMap and the reference model.
type orderedMapAPI interface {
	Set(key string, value int)
	Delete(key string)
	Iterate(f func(key string, value int) bool)
	Keys() []string
	Len() int
	Index(index int) (int, bool)
	CloneKeys() []string
	JSON() string
}

type realOrderedMap struct {
	omap *OrderedMap[string, int]
}

func (r realOrderedMap) Set(key string, value int)        { r.omap.Set(key, value) }
func (r realOrderedMap) Delete(key string)                { r.omap.Delete(key) }
func (r realOrderedMap) Iterate(f func(string, int) bool) { r.omap.Iterate(f) }
func (r realOrderedMap) Keys() []string                   { return orderedMapKeys(r.omap) }
func (r realOrderedMap) Len() int                         { return r.omap.Len() }
func (r realOrderedMap) Index(index int) (int, bool)      { return r.omap.GetByIndex(index) }
func (r realOrderedMap) JSON() string                     { return orderedMapJSON(r.omap) }

func (r realOrderedMap) CloneKeys() []string {
	clone := r.omap.Clone()

	return orderedMapKeys(&clone)
}

func orderedMapJSON(omap *OrderedMap[string, int]) string {
	data, err := json.Marshal(omap)
	if err != nil {
		return err.Error()
	}

	return string(data)
}

// modelOrderedMap removes keys immediately and iterates over a snapshot of the
// live keys, which is the behavior OrderedMap promises.
type modelOrderedMap struct {
	keys   []string
	values map[string]int
}

func (r *modelOrderedMap) Set(key string, value int) {
	if _, ok := r.values[key]; !ok {
		r.keys = append(r.keys, key)
	}

	r.values[key] = value
}

func (r *modelOrderedMap) Delete(key string) {
	if _, ok := r.values[key]; !ok {
		return
	}

	delete(r.values, key)

	r.keys = slices.DeleteFunc(r.keys, func(other string) bool { return other == key })
}

func (r *modelOrderedMap) Iterate(f func(string, int) bool) {
	for _, key := range slices.Clone(r.keys) {
		value, ok := r.values[key]
		if ok && !f(key, value) {
			return
		}
	}
}

func (r *modelOrderedMap) Keys() []string      { return slices.Clone(r.keys) }
func (r *modelOrderedMap) Len() int            { return len(r.keys) }
func (r *modelOrderedMap) CloneKeys() []string { return r.Keys() }

func (r *modelOrderedMap) Index(index int) (int, bool) {
	if index < 0 || index >= len(r.keys) {
		return 0, false
	}

	return r.values[r.keys[index]], true
}

func (r *modelOrderedMap) JSON() string {
	parts := make([]string, 0, len(r.keys))
	for _, key := range r.keys {
		parts = append(parts, fmt.Sprintf("%q:%d", key, r.values[key]))
	}

	return "{" + strings.Join(parts, ",") + "}"
}

// stepper is a small deterministic generator, so that failures replay from the
// seed alone.
type stepper int

func (s *stepper) intn(limit int) int {
	*s = (*s*1103515245 + 12345) & 0x7fffffff

	return int(*s>>8) % limit
}

// orderedMapDriver runs one random sequence of operations against an
// orderedMapAPI, including nested Iterate calls, breaks and reads inside
// callbacks, and records everything it observes.
type orderedMapDriver struct {
	omap   orderedMapAPI
	rng    stepper
	trace  []string
	budget int
	next   int
}

const (
	orderedMapDriverMaxDepth = 2
	orderedMapDriverKeys     = "abcd"
)

func (d *orderedMapDriver) logf(format string, args ...any) {
	d.trace = append(d.trace, fmt.Sprintf(format, args...))
}

func (d *orderedMapDriver) body(depth int) func(string, int) bool {
	return func(key string, value int) bool {
		d.logf("d%d visit %s=%d", depth, key, value)

		for range d.rng.intn(3) {
			if d.budget == 0 {
				break
			}

			d.budget--

			if !d.step(depth) {
				return false
			}
		}

		return true
	}
}

// step performs one random operation and reports whether the running
// iteration should continue.
func (d *orderedMapDriver) step(depth int) bool {
	target := string(orderedMapDriverKeys[d.rng.intn(len(orderedMapDriverKeys))])

	switch choice := d.rng.intn(10); {
	case choice < 3:
		d.next++
		d.omap.Set(target, d.next)
		d.logf("Set(%s,%d)", target, d.next)
	case choice < 6:
		d.omap.Delete(target)
		d.logf("Delete(%s)", target)
	case choice < 7 && depth < orderedMapDriverMaxDepth:
		d.logf("nested{")
		d.omap.Iterate(d.body(depth + 1))
		d.logf("}")
	case choice < 9:
		d.read()
	default:
		d.logf("break")

		return false
	}

	return true
}

func (d *orderedMapDriver) read() {
	index := d.rng.intn(5)
	got, ok := d.omap.Index(index)
	d.logf("keys=%v len=%d", d.omap.Keys(), d.omap.Len())
	d.logf("index(%d)=%d,%v clone=%v json=%s", index, got, ok, d.omap.CloneKeys(), d.omap.JSON())
}

func driveOrderedMap(omap orderedMapAPI, seed int) []string {
	driver := &orderedMapDriver{omap: omap, rng: stepper(seed), trace: nil, budget: 0, next: 100}
	driver.budget = 1 + driver.rng.intn(7)

	for idx := range driver.rng.intn(5) {
		omap.Set(string(orderedMapDriverKeys[idx%len(orderedMapDriverKeys)]), idx)
	}

	omap.Iterate(driver.body(0))
	driver.logf("final keys=%v len=%d json=%s", omap.Keys(), omap.Len(), omap.JSON())

	return driver.trace
}

func TestOrderedMapMatchesModel(t *testing.T) {
	t.Parallel()

	const sequences = 5000

	for seed := range sequences {
		omap := NewOrderedMap[string, int]()
		got := driveOrderedMap(realOrderedMap{omap: &omap}, seed)
		want := driveOrderedMap(&modelOrderedMap{keys: nil, values: map[string]int{}}, seed)

		require.Equal(t, want, got, "seed %d", seed)
	}
}
