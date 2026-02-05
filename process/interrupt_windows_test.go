//go:build windows

package process

import "testing"

func TestSendInterrupt(t *testing.T) {
	// On Windows CI (GitHub Actions), GenerateConsoleCtrlEvent may not work
	// as expected without a proper console attached.
	// This test verifies the function doesn't panic and the syscall loads correctly.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SendInterrupt panicked: %v", r)
		}
	}()

	// Just verify it doesn't crash - the actual signal delivery
	// depends on console configuration which varies in CI
	SendInterrupt()
}
