package healthcheck

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckConnection(t *testing.T) {
	t.Run("Test successful connection", func(t *testing.T) {
		info, err := CheckConnection("google.com", 80, "tcp", 1*time.Second)
		assert.NoError(t, err)
		assert.True(t, info.Successful)
		assert.Equal(t, "google.com", info.Host)
		assert.Contains(t, info.Message, "Successful")
	})

	t.Run("Test unsuccessful connection", func(t *testing.T) {
		_, err := CheckConnection("invalid.website", 80, "tcp", 1*time.Second)
		assert.Error(t, err)
	})

	t.Run("Test timeout connection", func(t *testing.T) {
		_, err := CheckConnection("192.0.2.0", 80, "tcp", 1*time.Millisecond)
		assert.Error(t, err)
	})
}
