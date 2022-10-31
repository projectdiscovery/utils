package executil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	// try to run the echo command
	s, err := Run("echo test")
	require.Nil(t, err, "failed execution", err)
	require.Equal(t, "test\n", s, "output doesn't contain expected result", s)
}

func TestRunSh(t *testing.T) {
	// try to run the echo command
	s, err := RunSh("echo", "test")
	require.Nil(t, err, "failed execution", err)
	require.Equal(t, "test\n", s, "output doesn't contain expected result", s)
}
