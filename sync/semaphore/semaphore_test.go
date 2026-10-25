package semaphore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVary(t *testing.T) {
	ctx := context.Background()
	s, err := New(2)
	require.NoError(t, err)

	require.NoError(t, s.Vary(ctx, 3))
	require.Equal(t, int64(5), s.Size())

	require.NoError(t, s.Vary(ctx, -4))
	require.Equal(t, int64(1), s.Size())

	require.NoError(t, s.Acquire(ctx, 1))
	require.False(t, s.sem.TryAcquire(1), "capacity should be exhausted after shrinking")
}
