package rawmode

import (
	"os"
	"syscall"
)

func Read(std *os.File, buf []byte) (int, error) {
	return syscall.Read(syscall.Handle(os.Stdin.Fd()), buf)
}
