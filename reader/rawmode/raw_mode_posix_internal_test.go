//go:build darwin || linux

package rawmode

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"
)

// ioctl issues request on file with arg as its argument, without going
// through the package under test. A nil arg passes no argument.
func ioctl[T any](file *os.File, request uintptr, arg *T) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), request, uintptr(unsafe.Pointer(arg)))
	if errno != 0 {
		return fmt.Errorf("ioctl %#x: %w", request, errno)
	}

	return nil
}

// swapStdin points os.Stdin at a pseudo-terminal that the code under test
// must leave alone, and returns the master and slave ends of that terminal.
// The decoy has echo turned off so that its attributes differ from those of
// a freshly opened pty, which lets a test tell the two apart. os.Stdin is
// process-global, so the tests that call it must not run in parallel.
func swapStdin(t *testing.T) (*os.File, *os.File) {
	t.Helper()

	decoyMaster, decoy := openPty(t)

	decoyTermios := termiosOf(t, decoy)
	decoyTermios.Lflag &^= syscall.ECHO
	setTermiosOf(t, decoy, decoyTermios)

	orig := os.Stdin

	t.Cleanup(func() { os.Stdin = orig })

	os.Stdin = decoy

	return decoyMaster, decoy
}

func TestGetModeLeavesTerminalUnchanged(t *testing.T) {
	_, slave := openPty(t)
	_, decoy := swapStdin(t)

	before := termiosOf(t, slave)
	decoyBefore := termiosOf(t, decoy)

	mode, err := GetMode(slave)
	if err != nil {
		t.Fatalf("GetMode: %v", err)
	}

	after := termiosOf(t, slave)
	if after != before {
		t.Errorf("GetMode changed the terminal:\nbefore %+v\nafter  %+v", before, after)
	}

	decoyAfter := termiosOf(t, decoy)
	if decoyAfter != decoyBefore {
		t.Errorf("GetMode touched os.Stdin instead of its argument")
	}

	got, ok := mode.(*syscall.Termios)
	if !ok {
		t.Fatalf("GetMode returned %T, want *syscall.Termios", mode)
	}

	if *got != before {
		t.Errorf("GetMode returned %+v, want the terminal's attributes %+v", *got, before)
	}

	if *got == decoyBefore {
		t.Errorf("GetMode returned the attributes of os.Stdin %+v", decoyBefore)
	}
}

func TestSetRawModeAndRestore(t *testing.T) {
	_, slave := openPty(t)
	_, decoy := swapStdin(t)

	const cooked = syscall.ECHO | syscall.ICANON

	original := termiosOf(t, slave)
	decoyBefore := termiosOf(t, decoy)

	if original.Lflag&cooked != cooked {
		t.Skipf("fresh pty is not in canonical mode with echo: lflag %#x", original.Lflag)
	}

	mode, err := GetMode(slave)
	if err != nil {
		t.Fatalf("GetMode: %v", err)
	}

	err = SetRawMode(slave, mode)
	if err != nil {
		t.Fatalf("SetRawMode: %v", err)
	}

	raw := termiosOf(t, slave)
	if raw.Lflag&cooked != 0 {
		t.Errorf("SetRawMode left echo or canonical mode on: lflag %#x", raw.Lflag)
	}

	restore := original

	err = SetMode(slave, &restore)
	if err != nil {
		t.Fatalf("SetMode: %v", err)
	}

	restored := termiosOf(t, slave)
	if restored != original {
		t.Errorf("SetMode did not restore the terminal:\nwant %+v\ngot  %+v", original, restored)
	}

	decoyAfter := termiosOf(t, decoy)
	if decoyAfter != decoyBefore {
		t.Errorf("mode changes touched os.Stdin instead of their argument")
	}
}

func TestReadUsesGivenFile(t *testing.T) {
	master, slave := openPty(t)
	decoyMaster, _ := swapStdin(t)

	// Without echo there is no unread output on the master, so closing the
	// slave at cleanup does not wait for it to drain.
	slaveTermios := termiosOf(t, slave)
	slaveTermios.Lflag &^= syscall.ECHO
	setTermiosOf(t, slave, slaveTermios)

	_, err := master.WriteString("given\n")
	if err != nil {
		t.Fatalf("write to pty master: %v", err)
	}

	// Pending input on os.Stdin keeps a Read of the wrong file from blocking.
	_, err = decoyMaster.WriteString("stdin\n")
	if err != nil {
		t.Fatalf("write to decoy pty master: %v", err)
	}

	buf := make([]byte, 16)

	n, err := Read(slave, buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	got, want := string(buf[:n]), "given\n"
	if got != want {
		t.Errorf("Read = %q, want %q", got, want)
	}
}

func TestGetModeNotATerminal(t *testing.T) {
	t.Parallel()

	plainFile, err := os.Create(filepath.Join(t.TempDir(), "plain"))
	if err != nil {
		t.Fatalf("create file: %v", err)
	}

	t.Cleanup(func() { _ = plainFile.Close() })

	mode, err := GetMode(plainFile)
	if !errors.Is(err, syscall.ENOTTY) {
		t.Errorf("GetMode error = %v, want ENOTTY", err)
	}

	if mode != nil {
		t.Errorf("GetMode returned %#v on failure, want a nil interface", mode)
	}

	err = SetRawMode(plainFile, mode)
	if err == nil {
		t.Error("SetRawMode accepted the mode returned by a failed GetMode")
	}
}
