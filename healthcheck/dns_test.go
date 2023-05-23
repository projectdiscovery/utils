package healthcheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDnsResolve(t *testing.T) {
	t.Run("Successful resolution", func(t *testing.T) {
		info, err := DnsResolve("scanme.sh", "1.1.1.1")
		assert.NoError(t, err)
		assert.True(t, info.Successful)
		assert.Equal(t, "scanme.sh", info.Host)
		assert.Equal(t, "1.1.1.1", info.Resolver)
		assert.NotEmpty(t, info.IPAddresses)
	})

	t.Run("Unsuccessful resolution due to invalid host", func(t *testing.T) {
		_, err := DnsResolve("invalid.website", "1.1.1.1")
		assert.Error(t, err)
	})

	t.Run("Unsuccessful resolution due to invalid resolver", func(t *testing.T) {
		_, err := DnsResolve("google.com", "invalid.resolver")
		assert.Error(t, err)
	})
}
