//go:build windows

package process

import (
	"os"
	"os/exec"
	"testing"
)

func TestSendInterrupt(t *testing.T) {
	// Re-exec the test in a child process so the CTRL_BREAK_EVENT does not
	// propagate to sibling processes (e.g. the Go compiler running in
	// parallel during "go test ./...").
	if os.Getenv("TEST_SEND_INTERRUPT_CHILD") == "1" {
		SendInterrupt()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSendInterrupt$")
	cmd.Env = append(os.Environ(), "TEST_SEND_INTERRUPT_CHILD=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child process failed: %v\n%s", err, out)
	}
}
