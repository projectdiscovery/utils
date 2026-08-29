//go:build windows

package process

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSendInterrupt(t *testing.T) {
	// Re-exec in a child with its own process group so the CTRL_BREAK_EVENT
	// stays isolated and does not kill sibling processes (e.g. the Go
	// compiler running in parallel during "go test ./...").
	if os.Getenv("TEST_SEND_INTERRUPT_CHILD") == "1" {
		// Register before raising: the runtime turns CTRL_BREAK_EVENT into
		// SIGINT only when something is watching for it, otherwise the console
		// default action terminates this process with STATUS_CONTROL_C_EXIT
		// (0xc000013a) and the parent sees a failed child.
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		defer signal.Stop(sigChan)

		SendInterrupt()

		select {
		case sig := <-sigChan:
			require.Equal(t, os.Interrupt, sig)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for interrupt signal")
		}
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
