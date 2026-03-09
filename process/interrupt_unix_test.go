//go:build !windows

package process

import (
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSendInterrupt(t *testing.T) {
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
}
