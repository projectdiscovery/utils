//go:build darwin || linux

package reader

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
)

// swapStdinWithFile points os.Stdin at a regular file, which is not a
// terminal. os.Stdin is process-global, so the tests that call it must not run
// in parallel.
func swapStdinWithFile(t *testing.T) {
	t.Helper()

	plainFile, err := os.Create(filepath.Join(t.TempDir(), "stdin"))
	if err != nil {
		t.Fatalf("create file: %v", err)
	}

	orig := os.Stdin

	t.Cleanup(func() {
		os.Stdin = orig

		_ = plainFile.Close()
	})

	os.Stdin = plainFile
}

func newRawKeyPressReader(once *sync.Once) *KeyPressReader {
	return &KeyPressReader{
		originalMode: nil,
		Timeout:      0,
		datachan:     nil,
		Once:         once,
		Raw:          true,
		BufferSize:   0,
	}
}

func TestKeyPressReaderStartRawNotATerminal(t *testing.T) {
	swapStdinWithFile(t)

	// A used Once keeps Start from launching the background read loop, which
	// is not what this test exercises.
	once := &sync.Once{}
	once.Do(func() {})

	keyPress := newRawKeyPressReader(once)

	err := keyPress.Start()

	_, ok := errors.AsType[syscall.Errno](err)
	if !ok {
		t.Fatalf("Start error = %v, want the errno from reading the terminal mode", err)
	}

	if !errors.Is(err, syscall.ENOTTY) {
		t.Errorf("Start error = %v, want ENOTTY", err)
	}
}

func TestKeyPressReaderStopWithoutMode(t *testing.T) {
	swapStdinWithFile(t)

	keyPress := newRawKeyPressReader(&sync.Once{})

	err := keyPress.Stop()
	if err != nil {
		t.Errorf("Stop = %v, want nil when Start never read a terminal mode", err)
	}
}
