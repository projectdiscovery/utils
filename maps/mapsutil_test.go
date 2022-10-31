package mapsutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeMaps(t *testing.T) {
	m1Str := map[string]interface{}{"a": 1, "b": 2}
	m2Str := map[string]interface{}{"b": 2, "c": 3}
	rStr := map[string]interface{}{"a": 1, "b": 2, "c": 3}
	rrStr := MergeMaps(m1Str, m2Str)
	require.EqualValues(t, rStr, rrStr)

	m1Int := map[int]interface{}{1: 1, 2: 2}
	m2Int := map[int]interface{}{1: 1, 2: 2, 3: 3, 4: 4}
	m3Int := map[int]interface{}{1: 1, 5: 5}
	rInt := map[int]interface{}{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	rrInt := MergeMaps(m1Int, m2Int, m3Int)
	require.EqualValues(t, rInt, rrInt)
}

func TestHTTPToMap(t *testing.T) {
	// not implemented
}

func TestDNSToMap(t *testing.T) {
	// not implemented
}

func TestHTTPRequesToMap(t *testing.T) {
	// not implemented
}

func TestHTTPResponseToMap(t *testing.T) {
	// not implemented
}

func TestGetKeys(t *testing.T) {
	t.Run("GetKeys(empty)", func(t *testing.T) {
		got := GetKeys(map[string]interface{}{})
		require.Empty(t, got)
	})

	t.Run("GetKeys(string)", func(t *testing.T) {
		got := GetKeys(map[string]interface{}{"a": "a", "b": "b"})
		require.EqualValues(t, []string{"a", "b"}, got)
	})

	t.Run("GetKeys(int)", func(t *testing.T) {
		got := GetKeys(map[int]interface{}{1: "a", 2: "b"})
		require.EqualValues(t, []int{1, 2}, got)
	})

	t.Run("GetKeys(bool)", func(t *testing.T) {
		got := GetKeys(map[bool]interface{}{true: "a", false: "b"})
		require.EqualValues(t, []bool{true, false}, got)
	})
}

func TestGetValues(t *testing.T) {
	t.Run("GetValues(empty)", func(t *testing.T) {
		got := GetValues(map[string]interface{}{})
		require.Empty(t, got)
	})

	t.Run("GetValues(string)", func(t *testing.T) {
		got := GetValues(map[string]interface{}{"a": "a", "b": "b"})
		require.EqualValues(t, []interface{}{"a", "b"}, got)
	})

	t.Run("GetValues(int)", func(t *testing.T) {
		got := GetValues(map[string]interface{}{"a": 1, "b": 2})
		require.EqualValues(t, []interface{}{1, 2}, got)
	})

	t.Run("GetValues(bool)", func(t *testing.T) {
		got := GetValues(map[string]interface{}{"a": true, "b": false})
		require.EqualValues(t, []interface{}{true, false}, got)
	})
}

func TestDifference(t *testing.T) {
	t.Run("Difference(empty)", func(t *testing.T) {
		got := Difference(map[string]interface{}{}, []string{}...)
		require.EqualValues(t, map[string]interface{}{}, got)
	})

	t.Run("Difference(string)", func(t *testing.T) {
		got := Difference(map[string]interface{}{"a": 1, "b": 2, "c": 3}, []string{"a"}...)
		require.EqualValues(t, map[string]interface{}{"b": 2, "c": 3}, got)
	})

	t.Run("Difference(int)", func(t *testing.T) {
		got := Difference(map[int]interface{}{1: "a", 2: "b", 3: "c"}, []int{1}...)
		require.EqualValues(t, map[int]interface{}{2: "b", 3: "c"}, got)
	})

	t.Run("Difference(bool)", func(t *testing.T) {
		got := Difference(map[bool]interface{}{true: 1, false: 2}, []bool{true}...)
		require.EqualValues(t, map[bool]interface{}{false: 2}, got)
	})
}
