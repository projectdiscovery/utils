package memoize

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemo(t *testing.T) {
	testingFunc := func() (interface{}, error) {
		time.Sleep(10 * time.Second)
		return "b", nil
	}

	m, err := New(WithMaxSize(5))
	require.Nil(t, err)
	start := time.Now()
	_, _, _ = m.Do("test", testingFunc)
	_, _, _ = m.Do("test", testingFunc)
	require.True(t, time.Since(start) < time.Duration(15*time.Second))
}

func TestSrc(t *testing.T) {
	out, err := File(PackageTemplate, "tests/test.go", "test")
	require.Nil(t, err)
	require.True(t, len(out) > 0)
}
