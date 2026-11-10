//go:build darwin

package rawmode

import (
	"bytes"
	"os"
	"syscall"
	"testing"
)

// openPty returns the master and slave ends of a new pseudo-terminal pair.
func openPty(t *testing.T) (*os.File, *os.File) {
	t.Helper()

	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("cannot open /dev/ptmx: %v", err)
	}

	t.Cleanup(func() { _ = master.Close() })

	err = ioctl[struct{}](master, syscall.TIOCPTYGRANT, nil)
	if err != nil {
		t.Skipf("grant pty slave: %v", err)
	}

	err = ioctl[struct{}](master, syscall.TIOCPTYUNLK, nil)
	if err != nil {
		t.Skipf("unlock pty slave: %v", err)
	}

	var name [128]byte

	err = ioctl(master, syscall.TIOCPTYGNAME, &name)
	if err != nil {
		t.Skipf("get pty slave name: %v", err)
	}

	path, _, _ := bytes.Cut(name[:], []byte{0})

	slave, err := os.OpenFile(string(path), os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("cannot open pty slave: %v", err)
	}

	t.Cleanup(func() { _ = slave.Close() })

	return master, slave
}

// termiosOf reads the terminal attributes of file without going through the
// package under test.
func termiosOf(t *testing.T, file *os.File) syscall.Termios {
	t.Helper()

	var term syscall.Termios

	err := ioctl(file, syscall.TIOCGETA, &term)
	if err != nil {
		t.Fatalf("get terminal attributes: %v", err)
	}

	return term
}

// setTermiosOf applies term to file without going through the package under
// test.
func setTermiosOf(t *testing.T, file *os.File, term syscall.Termios) {
	t.Helper()

	err := ioctl(file, syscall.TIOCSETA, &term)
	if err != nil {
		t.Fatalf("set terminal attributes: %v", err)
	}
}
