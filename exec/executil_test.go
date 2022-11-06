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
	// try to run the echo command
	s, err := Run("echo test")
	if runtime.GOOS == "windows" {
		require.NotNil(t, err)
		require.Empty(t, s)
	} else {
		require.Nil(t, err, "failed execution", err)
		require.Equal(t, "test"+newLineMarker, s, "output doesn't contain expected result", s)
	}
}

func TestRunSh(t *testing.T) {
	// try to run the echo command
	s, err := RunSh("echo", "test")
	if runtime.GOOS == "windows" {
		require.NotNil(t, err, "failed execution", err)
		require.Empty(t, s)
	} else {
		require.Nil(t, err, "failed execution", err)
		require.Equal(t, "test"+newLineMarker, s, "output doesn't contain expected result", s)
	}
}
