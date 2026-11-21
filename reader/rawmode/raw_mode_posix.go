//go:build darwin || linux

package rawmode

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func init() {
	GetMode = func(std *os.File) (interface{}, error) {
		termios, err := getMode(std)
		if err != nil {
			// Returning termios directly would box a nil *syscall.Termios into
			// a non-nil interface, which SetMode and SetRawMode would accept.
			return nil, err
		}

		return termios, nil
	}

	SetMode = func(std *os.File, mode interface{}) error {
		m, ok := mode.(*syscall.Termios)
		if !ok {
			return errors.New("invalid syscall.Termios")
		}
		return setMode(std, m)
	}

	SetRawMode = func(std *os.File, mode interface{}) error {
		m, ok := mode.(*syscall.Termios)
		if !ok {
			return errors.New("invalid syscall.Termios")
		}
		return setRawMode(std, m)
	}

	Read = func(std *os.File, buf []byte) (int, error) {
		return read(std, buf)
	}
}

func getTermios(fd uintptr) (*syscall.Termios, error) {
	var t syscall.Termios
	// A raw syscall reports success as Errno(0), which is a non-nil error
	// once converted to the error interface.
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		fd,
		TCGETS,
		uintptr(unsafe.Pointer(&t)),
		0, 0, 0)
	if errno != 0 {
		return nil, fmt.Errorf("get terminal attributes: %w", errno)
	}

	return &t, nil
}

func setTermios(fd uintptr, term *syscall.Termios) error {
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		fd,
		TCSETS,
		uintptr(unsafe.Pointer(term)),
		0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("set terminal attributes: %w", errno)
	}

	return nil
}

func setRaw(term *syscall.Termios) {
	// This attempts to replicate the behaviour documented for cfmakeraw in
	// the termios(3) manpage.
	term.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	term.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	term.Cflag &^= syscall.CSIZE | syscall.PARENB
	term.Cflag |= syscall.CS8

	term.Cc[syscall.VMIN] = 1
	term.Cc[syscall.VTIME] = 0
}

func getMode(std *os.File) (*syscall.Termios, error) {
	return getTermios(std.Fd())
}

func setMode(std *os.File, mode *syscall.Termios) error {
	return setTermios(std.Fd(), mode)
}

func setRawMode(std *os.File, mode *syscall.Termios) error {
	setRaw(mode)
	return SetMode(std, mode)
}

func read(std *os.File, buf []byte) (int, error) {
	n, err := syscall.Read(int(std.Fd()), buf)
	if err != nil {
		return n, fmt.Errorf("read from file: %w", err)
	}

	return n, nil
}
