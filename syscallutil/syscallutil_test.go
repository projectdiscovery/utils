package syscallutil

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadLibrary(t *testing.T) {
	t.Run("Test valid library", func(t *testing.T) {
		var lib string
		if runtime.GOOS == "windows" {
			lib = "ucrtbase.dll"
		} else {
			lib = "libc.so.6"
		}

		_, err := LoadLibrary(lib)
		require.NoError(t, err, "should not return an error for valid library")
	})

	t.Run("Test invalid library", func(t *testing.T) {
		var lib string
		if runtime.GOOS == "windows" {
			lib = "C:\\path\\to\\invalid\\library.dll"
		} else {
			lib = "/path/to/invalid/library.so"
		}

		_, err := LoadLibrary(lib)
		require.Error(t, err, "should return an error for invalid library")
	})
}
