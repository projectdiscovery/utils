package generic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEqualsAnyInt(t *testing.T) {
	testCases := []struct {
		Base     int
		All      []int
		Expected bool
	}{
		{3, []int{1, 2, 3, 4}, true},
		{5, []int{1, 2, 3, 4}, false},
		{0, []int{0}, true},
		{0, []int{1}, false},
	}

	for _, tc := range testCases {
		actual := EqualsAny(tc.Base, tc.All...)
		require.Equal(t, tc.Expected, actual)
	}
}

func TestEqualsAnyString(t *testing.T) {
	testCases := []struct {
		Base     string
		All      []string
		Expected bool
	}{
		{"test", []string{"test1", "test", "test2", "test3"}, true},
		{"test", []string{"test1", "test2", "test3", "test4"}, false},
		{"", []string{""}, true},
		{"", []string{"not empty"}, false},
	}

	for _, tc := range testCases {
		actual := EqualsAny(tc.Base, tc.All...)
		require.Equal(t, tc.Expected, actual)
	}
}
