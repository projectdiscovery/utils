package sysutil

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetMaxThreads(t *testing.T) {
	originalMaxThreads := debug.SetMaxThreads(10000)
	defer debug.SetMaxThreads(originalMaxThreads)

	newMaxThreads := 5000
	previousMaxThreads := SetMaxThreads(newMaxThreads)
	require.Equal(t, 10000, previousMaxThreads, "Expected previous max threads to be 10000")
	require.Equal(t, newMaxThreads, debug.SetMaxThreads(newMaxThreads), "Expected max threads to be set to 5000")

	SetMaxThreads(originalMaxThreads)
}
