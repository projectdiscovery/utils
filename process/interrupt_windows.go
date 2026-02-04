//go:build windows

package process

import "syscall"

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procGenerateConsoleCtrlEvent = kernel32.NewProc("GenerateConsoleCtrlEvent")
)

// SendInterrupt sends an interrupt signal to the current process.
// On Windows, this uses GenerateConsoleCtrlEvent with CTRL_BREAK_EVENT
// because Go's p.Signal(os.Interrupt) is not implemented on Windows.
func SendInterrupt() {
	// CTRL_BREAK_EVENT = 1
	procGenerateConsoleCtrlEvent.Call(1, 0)
}
