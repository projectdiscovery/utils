//go:build darwin || linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

func getTermios(fd uintptr) (*syscall.Termios, error) {
	var t syscall.Termios
	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		syscall.TCGETS,
		uintptr(unsafe.Pointer(&t)),
		0, 0, 0)

	return &t, err
}

func setTermios(fd uintptr, term *syscall.Termios) error {
	_, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		os.Stdin.Fd(),
		syscall.TCSETS,
		uintptr(unsafe.Pointer(term)),
		0, 0, 0)
	return err
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

func GetMode(std *os.File) (*syscall.Termios, error) {
	return getTermios(os.Stdin.Fd())
}

func SetMode(std *os.File, mode *syscall.Termios) error {
	return setTermMode(os.Stdin.Fd(), mode)
}

func SetRawMode(std *os.File, mode *syscall.Termios) error {
	setRaw(mode)
	return SetMode(std, mode)
}
