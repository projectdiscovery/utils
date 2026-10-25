package connpool

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOneTimePoolAcquirePrefersCanceledContext(t *testing.T) {
	for range 100 {
		ctx, cancel := context.WithCancel(context.Background())
		pool, err := NewOneTimePool(ctx, "unused", 1)
		require.NoError(t, err)

		client, server := net.Pipe()
		pool.idleConnections <- client
		cancel()

		conn, err := pool.Acquire(context.Background())
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, conn)

		_ = client.Close()
		_ = server.Close()
		_ = pool.Close()
	}
}
