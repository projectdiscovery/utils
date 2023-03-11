package mapsutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapHas(t *testing.T) {
	m := Map[string, int]{"foo": 1, "bar": 2}
	testCases := []struct {
		key      string
		expected bool
	}{
		{"foo", true},
		{"baz", false},
	}
	for _, tc := range testCases {
		actual := m.Has(tc.key)
		if actual != tc.expected {
			t.Errorf("Has(%q) = %v, expected %v", tc.key, actual, tc.expected)
		}
	}
}

func TestMapGetKeys(t *testing.T) {
	m := Map[string, int]{"foo": 1, "bar": 2}
	testCases := []struct {
		keys     []string
		expected []int
	}{
		{[]string{"foo", "bar"}, []int{1, 2}},
		{[]string{"baz", "qux"}, []int{0, 0}},
	}
	for _, tc := range testCases {
		actual := m.GetKeys(tc.keys...)
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("GetKeys(%v) = %v, expected %v", tc.keys, actual, tc.expected)
		}
	}
}

func TestMapGetOrDefault(t *testing.T) {
	m := Map[string, int]{"foo": 1, "bar": 2}
	testCases := []struct {
		key      string
		defaultV int
		expected int
	}{
		{"foo", 0, 1},
		{"baz", 0, 0},
	}
	for _, tc := range testCases {
		actual := m.GetOrDefault(tc.key, tc.defaultV)
		if actual != tc.expected {
			t.Errorf("GetOrDefault(%q, %d) = %d, expected %d", tc.key, tc.defaultV, actual, tc.expected)
		}
	}
}

func TestMapMerge(t *testing.T) {
	m := Map[string, int]{"foo": 1, "bar": 2}
	n := map[string]int{"baz": 3, "qux": 4}
	m.Merge(n)
	expected := Map[string, int]{"foo": 1, "bar": 2, "baz": 3, "qux": 4}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("Merge(%v) = %v, expected %v", n, m, expected)
	}
}

func TestMap_GetKeyWithValue(t *testing.T) {
	type testCase[K, V comparable] struct {
		InputMap    Map[K, V]
		Value       V
		ExpectedKey K
		ExpectedOk  bool
	}

	genericMap := Map[string, string]{"a": "a", "b": "b", "c": "c"}

	testCases := []testCase[string, string]{
		{
			InputMap:    genericMap,
			Value:       "b",
			ExpectedKey: "b",
			ExpectedOk:  true,
		},
		{
			InputMap:    genericMap,
			Value:       "d",
			ExpectedKey: "",
			ExpectedOk:  false,
		},
		{
			InputMap:    genericMap,
			Value:       "b",
			ExpectedKey: "b",
			ExpectedOk:  true,
		},
		{
			InputMap:    genericMap,
			Value:       "d",
			ExpectedKey: "",
			ExpectedOk:  false,
		},
		{
			InputMap:    genericMap,
			Value:       "value",
			ExpectedKey: "",
			ExpectedOk:  false,
		},
	}

	for _, tc := range testCases {
		key, ok := tc.InputMap.GetKeyWithValue(tc.Value)
		require.Equal(t, tc.ExpectedKey, key)
		require.Equal(t, tc.ExpectedOk, ok)
	}
}
