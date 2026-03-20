//go:build windows

package process

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
)

func TestSendInterrupt(t *testing.T) {
	// Re-exec in a child with its own process group so the CTRL_BREAK_EVENT
	// stays isolated and does not kill sibling processes (e.g. the Go
	// compiler running in parallel during "go test ./...").
	if os.Getenv("TEST_SEND_INTERRUPT_CHILD") == "1" {
		SendInterrupt()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSendInterrupt$")
	cmd.Env = append(os.Environ(), "TEST_SEND_INTERRUPT_CHILD=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child process failed: %v\n%s", err, out)
	}
}
