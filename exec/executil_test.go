package executil

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

var newLineMarker string

func init() {
	if runtime.GOOS == "windows" {
		newLineMarker = "\r\n"
	} else {
		newLineMarker = "\n"
	}
}

func TestRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	// try to run the echo command
	s, err := Run("echo test")
	require.Nil(t, err, "failed execution", err)
	require.Equal(t, "test"+newLineMarker, s, "output doesn't contain expected result", s)
}

func TestRunSh(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
	// try to run the echo command
	s, err := RunSh("echo", "test")
	require.Nil(t, err, "failed execution", err)
	require.Equal(t, "test"+newLineMarker, s, "output doesn't contain expected result", s)
}
