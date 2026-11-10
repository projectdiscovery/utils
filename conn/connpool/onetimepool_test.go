package connpool

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

// waitRunTimeout bounds how long a test waits for Run to return.
const waitRunTimeout = 5 * time.Second

var errRefused = errors.New("connection refused")

// funcDialer adapts a function to the Dialer interface.
type funcDialer func(ctx context.Context, network, address string) (net.Conn, error)

func (f funcDialer) Dial(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

// connTracker closes every connection handed to it when the test ends, so
// fake dialers and listeners never leak the peer side of a connection.
type connTracker struct {
	mu     sync.Mutex
	conns  []net.Conn
	closed bool
}

func newConnTracker(t *testing.T) *connTracker {
	t.Helper()

	tracker := new(connTracker)
	t.Cleanup(tracker.closeAll)

	return tracker
}

func (c *connTracker) add(conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		_ = conn.Close()

		return
	}

	c.conns = append(c.conns, conn)
}

func (c *connTracker) first() net.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conns[0]
}

func (c *connTracker) closeAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}

// pipe returns the client side of a new in-memory connection and tracks the
// server side.
func (c *connTracker) pipe() net.Conn {
	client, server := net.Pipe()
	c.add(server)

	return client
}

func newTestPool(t *testing.T, dialer Dialer, size int) *OneTimePool {
	t.Helper()

	pool, err := NewOneTimePool(context.Background(), "unused", size, WithDialer(dialer))
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })

	return pool
}

// runPool starts pool.Run in a goroutine and returns a channel that receives
// its result.
func runPool(pool *OneTimePool) <-chan error {
	done := make(chan error, 1)
	go func() { done <- pool.Run() }()

	return done
}

func waitRun(t *testing.T, done <-chan error) error {
	t.Helper()

	select {
	case err := <-done:
		return err
	case <-time.After(waitRunTimeout):
		t.Fatalf("Run did not return within %s", waitRunTimeout)

		return nil
	}
}

// waitAfterFailure is the wait Run applies after the nth consecutive failed
// dial (1-based), derived from the backoff constants.
func waitAfterFailure(n int) time.Duration {
	wait := minDialBackoff
	for range n - 1 {
		wait = min(wait*backoffFactor, maxDialBackoff)
	}

	return wait
}

// maxDials is the most dial attempts Run can make in the given time when
// every dial fails. Timers never fire early, so attempt n+1 cannot start
// before the first n waits have elapsed.
func maxDials(elapsed time.Duration) int64 {
	dials := int64(1)

	var waited time.Duration

	for n := 1; ; n++ {
		waited += waitAfterFailure(n)
		if waited > elapsed {
			return dials
		}

		dials++
	}
}

func TestOneTimePoolRunBacksOffOnDialErrors(t *testing.T) {
	t.Parallel()

	var dials atomic.Int64

	pool := newTestPool(t, funcDialer(func(context.Context, string, string) (net.Conn, error) {
		dials.Add(1)

		return nil, errRefused
	}), 1)

	start := time.Now()
	done := runPool(pool)

	time.Sleep(400 * time.Millisecond)
	require.NoError(t, pool.Close())
	require.ErrorIs(t, waitRun(t, done), context.Canceled)

	// A tight retry loop with no wait between dials would blow far past the
	// schedule-derived bound.
	require.GreaterOrEqual(t, dials.Load(), int64(1))
	require.LessOrEqual(t, dials.Load(), maxDials(time.Since(start)))
}

func TestOneTimePoolCloseInterruptsBackoff(t *testing.T) {
	t.Parallel()

	// The wait after this many failures is the longest one still below the cap.
	const failures = 5

	var dials atomic.Int64

	pool := newTestPool(t, funcDialer(func(context.Context, string, string) (net.Conn, error) {
		dials.Add(1)

		return nil, errRefused
	}), 1)
	done := runPool(pool)

	require.Eventually(t, func() bool { return dials.Load() >= failures }, 10*time.Second, time.Millisecond)

	// Run is now parked in a wait of about pending. An uninterruptible wait
	// would delay its return by most of that.
	pending := waitAfterFailure(failures)
	closeStart := time.Now()

	require.NoError(t, pool.Close())
	require.ErrorIs(t, waitRun(t, done), context.Canceled)
	require.Less(t, time.Since(closeStart), pending/2)
}

func TestOneTimePoolBackoffResetsAfterSuccess(t *testing.T) {
	t.Parallel()

	// Failures before the one success. Without a reset, the wait after the
	// next failure would be waitAfterFailure(failures+1).
	const failures = 4

	var (
		stampsMu sync.Mutex
		stamps   []time.Time
	)

	tracker := newConnTracker(t)
	pool := newTestPool(t, funcDialer(func(context.Context, string, string) (net.Conn, error) {
		stampsMu.Lock()
		defer stampsMu.Unlock()

		stamps = append(stamps, time.Now())
		if len(stamps) == failures+1 {
			return tracker.pipe(), nil
		}

		return nil, errRefused
	}), 1)
	done := runPool(pool)

	// The pool holds one connection, so after the success Run dials again,
	// fails, waits, and fails once more.
	require.Eventually(t, func() bool {
		stampsMu.Lock()
		defer stampsMu.Unlock()

		return len(stamps) >= failures+3
	}, 10*time.Second, time.Millisecond)
	require.NoError(t, pool.Close())
	require.ErrorIs(t, waitRun(t, done), context.Canceled)

	stampsMu.Lock()
	defer stampsMu.Unlock()

	gap := stamps[failures+2].Sub(stamps[failures+1])
	require.GreaterOrEqual(t, gap, minDialBackoff)
	require.Less(t, gap, waitAfterFailure(failures+1)/2)
}

func TestOneTimePoolRecoversAfterDialErrors(t *testing.T) {
	t.Parallel()

	var dials atomic.Int64

	tracker := newConnTracker(t)
	pool := newTestPool(t, funcDialer(func(context.Context, string, string) (net.Conn, error) {
		if dials.Add(1) <= 2 {
			return nil, errRefused
		}

		return tracker.pipe(), nil
	}), 1)
	done := runPool(pool)

	ctx, cancel := context.WithTimeout(context.Background(), waitRunTimeout)
	defer cancel()

	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.GreaterOrEqual(t, dials.Load(), int64(3))

	go func() { _, _ = conn.Write([]byte("ping")) }()

	buf := make([]byte, len("ping"))
	_, err = io.ReadFull(tracker.first(), buf)
	require.NoError(t, err)
	require.Equal(t, "ping", string(buf))

	_ = conn.Close()

	require.NoError(t, pool.Close())
	require.ErrorIs(t, waitRun(t, done), context.Canceled)
}

func TestOneTimePoolRunFillsPool(t *testing.T) {
	t.Parallel()

	tracker := newConnTracker(t)
	pool := newTestPool(t, funcDialer(func(context.Context, string, string) (net.Conn, error) {
		return tracker.pipe(), nil
	}), 2)
	done := runPool(pool)

	ctx, cancel := context.WithTimeout(context.Background(), waitRunTimeout)
	defer cancel()

	// More acquisitions than the pool size proves Run keeps refilling it.
	for range 5 {
		conn, err := pool.Acquire(ctx)
		require.NoError(t, err)
		require.NotNil(t, conn)

		_ = conn.Close()
	}

	require.NoError(t, pool.Close())
	require.ErrorIs(t, waitRun(t, done), context.Canceled)
}

func TestOneTimePoolRunWithoutDialer(t *testing.T) {
	t.Parallel()

	var lc net.ListenConfig

	listener, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	tracker := newConnTracker(t)
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		for {
			accepted, err := listener.Accept()
			if err != nil {
				return
			}

			tracker.add(accepted)
		}
	}()

	pool, err := NewOneTimePool(context.Background(), listener.Addr().String(), 1)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })

	done := runPool(pool)

	ctx, cancel := context.WithTimeout(context.Background(), waitRunTimeout)
	defer cancel()

	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	require.NotNil(t, conn)

	_ = conn.Close()

	require.NoError(t, pool.Close())
	require.ErrorIs(t, waitRun(t, done), context.Canceled)
}

// Close must not wait for a dial in progress, and a connection that finishes
// dialing after Close must not be left open.
func TestOneTimePoolCloseDuringDial(t *testing.T) {
	t.Parallel()

	var (
		entered = make(chan struct{})
		release = make(chan struct{})
		once    sync.Once
	)

	tracker := newConnTracker(t)
	dialed := tracker.pipe()
	server := tracker.first()
	pool := newTestPool(t, funcDialer(func(context.Context, string, string) (net.Conn, error) {
		once.Do(func() { close(entered) })
		<-release

		return dialed, nil
	}), 1)
	done := runPool(pool)

	<-entered

	closed := make(chan error, 1)
	go func() { closed <- pool.Close() }()

	select {
	case err := <-closed:
		require.NoError(t, err)
	case <-time.After(time.Second):
		close(release)
		t.Fatal("Close blocked while a dial was in progress")
	}

	close(release)
	require.ErrorIs(t, waitRun(t, done), context.Canceled)

	// The dialed connection was discarded, so the peer observes EOF.
	readErr := make(chan error, 1)
	go func() {
		_, err := server.Read(make([]byte, 1))
		readErr <- err
	}()

	select {
	case err := <-readErr:
		require.ErrorIs(t, err, io.EOF)
	case <-time.After(waitRunTimeout):
		t.Fatal("dialed connection was left open after Close")
	}
}
