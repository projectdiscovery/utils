package sliceutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPruneEmptyStrings(t *testing.T) {
	test := []string{"a", "", "", "b"}
	// converts back
	res := PruneEmptyStrings(test)
	require.Equal(t, []string{"a", "b"}, res, "strings not pruned correctly")
}

func TestPruneEqual(t *testing.T) {
	testStr := []string{"a", "", "", "b"}
	// converts back
	resStr := PruneEqual(testStr, "b")
	require.Equal(t, []string{"a", "", ""}, resStr, "strings not pruned correctly")

	testInt := []int{1, 2, 3, 4}
	// converts back
	resInt := PruneEqual(testInt, 2)
	require.Equal(t, []int{1, 3, 4}, resInt, "ints not pruned correctly")
}

func TestDedupe(t *testing.T) {
	testStr := []string{"a", "a", "b", "b"}
	// converts back
	resStr := Dedupe(testStr)
	require.Equal(t, []string{"a", "b"}, resStr, "strings not deduped correctly")

	testInt := []int{1, 1, 2, 2}
	// converts back
	res := Dedupe(testInt)
	require.Equal(t, []int{1, 2}, res, "ints not deduped correctly")
}

func TestPickRandom(t *testing.T) {
	testStr := []string{"a", "b"}
	// converts back
	resStr := PickRandom(testStr)
	require.Contains(t, testStr, resStr, "element was not picked correctly")

	testInt := []int{1, 2}
	// converts back
	resInt := PickRandom(testInt)
	require.Contains(t, testInt, resInt, "element was not picked correctly")
}

func TestContains(t *testing.T) {
	testSliceStr := []string{"a", "b"}
	testElemStr := "a"
	// converts back
	resStr := Contains(testSliceStr, testElemStr)
	require.True(t, resStr, "unexptected result")

	testSliceInt := []int{1, 2}
	testElemInt := 1
	// converts back
	resInt := Contains(testSliceInt, testElemInt)
	require.True(t, resInt, "unexptected result")
}

func TestContainsItems(t *testing.T) {
	test1Str := []string{"a", "b", "c"}
	test2Str := []string{"a", "c"}
	// converts back
	resStr := ContainsItems(test1Str, test2Str)
	require.True(t, resStr, "unexptected result")

	test1Int := []int{1, 2, 3}
	test2Int := []int{1, 3}
	// converts back
	resInt := ContainsItems(test1Int, test2Int)
	require.True(t, resInt, "unexptected result")
}

func TestToInt(t *testing.T) {
	test1 := []string{"1", "2"}
	test2 := []int{1, 2}
	// converts back
	res, err := ToInt(test1)
	require.Nil(t, err)
	require.Equal(t, test2, res, "unexptected result")
}
