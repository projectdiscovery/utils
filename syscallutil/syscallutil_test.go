package syscallutil

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadLibrary(t *testing.T) {
	t.Run("Test valid library", func(t *testing.T) {
		var lib string
		switch runtime.GOOS {
		case "windows":
			lib = "ucrtbase.dll"
		case "darwin":
			lib = "libSystem.dylib"
		case "linux":
			lib = "libc.so.6"
		default:
			panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
		}

		_, err := LoadLibrary(lib)
		require.NoError(t, err, "should not return an error for valid library")
	})

	t.Run("Test invalid library", func(t *testing.T) {
		var lib string
		switch runtime.GOOS {
		case "windows":
			lib = "C:\\path\\to\\invalid\\library.dll"
		case "darwin":
			lib = "/path/to/invalid/library.dylib"
		case "linux":
			lib = "/path/to/invalid/library.so"
		default:
			panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
		}

		_, err := LoadLibrary(lib)
		require.Error(t, err, "should return an error for invalid library")
	})
}
