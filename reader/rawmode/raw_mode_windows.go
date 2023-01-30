//go:build windows

package rawmode

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	// load kernel32 lib
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// get handlers to console API
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

const (
	enableLineInput       = 2
	enableEchoInput       = 4
	enableProcessedInput  = 1
	enableWindowInput     = 8
	enableMouseInput      = 16
	enableInsertMode      = 32
	enableQuickEditMode   = 64
	enableExtendedFlags   = 128
	enableAutoPosition    = 256
	enableProcessedOutput = 1
	enableWrapAtEolOutput = 2
)

func getTermMode(fd uintptr) (uint32, error) {
	var mode uint32
	_, _, err := syscall.Syscall(
		procGetConsoleMode.Addr(),
		2,
		fd,
		uintptr(unsafe.Pointer(&mode)),
		0)
	if err != 0 {
		return mode, err
	}
	return mode, nil
}

func setTermMode(fd uintptr, mode uint32) error {
	_, _, err := syscall.Syscall(
		procSetConsoleMode.Addr(),
		2,
		fd,
		uintptr(mode),
		0)
	if err != 0 {
		return err
	}
	return nil
}

// GetMode from file descriptor
func GetMode(std *os.File) (uint32, error) {
	return getTermMode(os.Stdin.Fd())
}

// SetMode to file descriptor
func SetMode(std *os.File, mode uint32) error {
	return setTermMode(os.Stdin.Fd(), mode)
}

// SetRawMode to file descriptor enriching existign mode with raw console flags
func SetRawMode(std *os.File, mode uint32) error {
	mode &^= (enableEchoInput | enableProcessedInput | enableLineInput | enableProcessedOutput)
	return SetMode(std, mode)
}

// Read from file descriptor to buffer
func Read(std *os.File, buf []byte) (int, error) {
	return syscall.Read(syscall.Handle(os.Stdin.Fd()), buf)
}
