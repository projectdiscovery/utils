package reader

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const lateData = "LATE"

// gatedReader blocks every Read until release is closed, then delivers
// lateData. entered is signalled when a Read starts, done when it has
// finished writing into the buffer it was given.
type gatedReader struct {
	release chan struct{}
	entered chan struct{}
	done    chan struct{}
}

func newGatedReader() *gatedReader {
	return &gatedReader{
		release: make(chan struct{}),
		entered: make(chan struct{}, 8),
		done:    make(chan struct{}, 8),
	}
}

func (g *gatedReader) Read(p []byte) (int, error) {
	g.entered <- struct{}{}

	<-g.release

	n := copy(p, lateData)

	g.done <- struct{}{}

	return n, nil
}

// readResult carries the values of one Read call. The error is stored as the
// source returned it, so tests can assert its identity.
type readResult struct {
	count int
	err   error
}

// readNoPanic turns a panic in Read into a test failure so one broken case
// cannot take down the rest of the package's tests.
func readNoPanic(t *testing.T, src io.Reader, buf []byte) readResult {
	t.Helper()

	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Read panicked: %v", rec)
		}
	}()

	count, err := src.Read(buf)

	return readResult{count: count, err: err}
}

func waitSignal(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func TestTimeoutReader(t *testing.T) {
	wrappedStdin := TimeoutReader{
		Reader:  FrozenReader{},
		Timeout: time.Duration(2 * time.Second),
	}
	_, err := io.Copy(os.Stdout, wrappedStdin)
	require.NotNil(t, err)
}

func TestTimeoutReaderTimeoutResult(t *testing.T) {
	t.Parallel()

	src := newGatedReader()
	defer close(src.release)

	res := readNoPanic(t, TimeoutReader{Reader: src, Timeout: 20 * time.Millisecond}, make([]byte, 4))
	require.Equal(t, 0, res.count)
	require.Same(t, ErrTimeout, res.err)
}

func TestTimeoutReaderLateBytesDoNotTouchCallerBuffer(t *testing.T) {
	t.Parallel()

	src := newGatedReader()
	buf := make([]byte, 4)

	res := readNoPanic(t, TimeoutReader{Reader: src, Timeout: 20 * time.Millisecond}, buf)
	require.ErrorIs(t, res.err, ErrTimeout)

	// The caller owns buf again once Read has returned.
	copy(buf, "MINE")
	close(src.release)
	waitSignal(t, src.done, "source to finish its late read")

	require.Equal(t, "MINE", string(buf))
}

// The source here is released by the clock rather than by a channel, so
// nothing orders its late write against the caller's reuse of buf; the race
// detector reports an unsynchronised access if Read leaves a writer behind.
// Any synchronisation between the caller's write and the source's would hide
// that, so the test relies on a wide ratio between the source delay and the
// timeout instead.
func TestTimeoutReaderNoRaceOnCallerBuffer(t *testing.T) {
	t.Parallel()

	finished := make(chan struct{})
	src := &delayedReader{delay: 200 * time.Millisecond, data: lateData, err: nil, finished: finished}
	buf := make([]byte, 4)

	res := readNoPanic(t, TimeoutReader{Reader: src, Timeout: 10 * time.Millisecond}, buf)
	require.ErrorIs(t, res.err, ErrTimeout)

	copy(buf, "MINE")
	waitSignal(t, finished, "source to finish its late read")

	require.Equal(t, "MINE", string(buf))
}

// Stays sequential: it compares process-wide goroutine counts, and Go only
// runs a non-parallel test while parallel tests are paused, so no other test's
// helper goroutines can skew the count.
func TestTimeoutReaderHelperGoroutineExits(t *testing.T) {
	baseline := runtime.NumGoroutine()
	src := newGatedReader()

	res := readNoPanic(t, TimeoutReader{Reader: src, Timeout: 20 * time.Millisecond}, make([]byte, 4))
	require.ErrorIs(t, res.err, ErrTimeout)
	waitSignal(t, src.entered, "source Read to start")

	close(src.release)
	waitSignal(t, src.done, "source to finish its late read")

	deadline := time.Now().Add(5 * time.Second)
	for runtime.NumGoroutine() > baseline && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}

	require.LessOrEqual(t, runtime.NumGoroutine(), baseline, "helper goroutine still running after the source returned")
}

// Bytes the source delivers after a timeout are dropped, not replayed on the
// next Read.
func TestTimeoutReaderLateBytesAreDropped(t *testing.T) {
	t.Parallel()

	gate := newGatedReader()
	source := &gateThenNextReader{gate: gate, called: atomic.Bool{}}
	timeoutReader := TimeoutReader{Reader: source, Timeout: 20 * time.Millisecond}

	res := readNoPanic(t, timeoutReader, make([]byte, 4))
	require.ErrorIs(t, res.err, ErrTimeout)
	close(gate.release)
	waitSignal(t, gate.done, "source to finish its late read")

	buf := make([]byte, 4)
	res = readNoPanic(t, timeoutReader, buf)
	require.NoError(t, res.err)
	require.Equal(t, nextData, string(buf[:res.count]))
}

func TestTimeoutReaderNonPositiveTimeoutReadsWithoutTimeout(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		timeout time.Duration
	}{
		{name: "zero", timeout: 0},
		{name: "negative", timeout: -time.Second},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			slow := &delayedReader{delay: 30 * time.Millisecond, data: "slow", err: nil, finished: nil}
			buf := make([]byte, 8)

			res := readNoPanic(t, TimeoutReader{Reader: slow, Timeout: testCase.timeout}, buf)
			require.NoError(t, res.err)
			require.Equal(t, "slow", string(buf[:res.count]))
		})
	}
}

func TestTimeoutReaderFastRead(t *testing.T) {
	t.Parallel()

	timeoutReader := TimeoutReader{Reader: bytes.NewReader([]byte("hello")), Timeout: time.Second}
	buf := make([]byte, 3)

	res := readNoPanic(t, timeoutReader, buf)
	require.NoError(t, res.err)
	require.Equal(t, "hel", string(buf[:res.count]))

	res = readNoPanic(t, timeoutReader, buf)
	require.NoError(t, res.err)
	require.Equal(t, "lo", string(buf[:res.count]))

	res = readNoPanic(t, timeoutReader, buf)
	require.Equal(t, 0, res.count)
	require.Same(t, io.EOF, res.err)
}

func TestTimeoutReaderPassesThroughDataWithError(t *testing.T) {
	t.Parallel()

	wantErr := io.ErrUnexpectedEOF
	src := &delayedReader{delay: 0, data: "tail", err: wantErr, finished: nil}
	buf := make([]byte, 8)

	res := readNoPanic(t, TimeoutReader{Reader: src, Timeout: time.Second}, buf)
	require.Equal(t, "tail", string(buf[:res.count]))
	require.Same(t, wantErr, res.err)
}

func TestTimeoutReaderReadAll(t *testing.T) {
	t.Parallel()

	want := bytes.Repeat([]byte("0123456789"), 5000)
	got, err := io.ReadAll(TimeoutReader{Reader: bytes.NewReader(want), Timeout: time.Second})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func BenchmarkTimeoutReader(b *testing.B) {
	for _, size := range []int{512, 32 * 1024} {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			timeoutReader := TimeoutReader{Reader: infiniteReader{}, Timeout: time.Minute}
			buf := make([]byte, size)

			b.ReportAllocs()
			b.SetBytes(int64(size))

			for b.Loop() {
				_, err := timeoutReader.Read(buf)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

type infiniteReader struct{}

func (infiniteReader) Read(p []byte) (int, error) { return len(p), nil }

// delayedReader returns data (and err) after sleeping for delay.
type delayedReader struct {
	delay time.Duration
	data  string
	err   error
	// finished, when set, is closed after the data has been written.
	finished chan struct{}
}

func (d *delayedReader) Read(p []byte) (int, error) {
	time.Sleep(d.delay)

	count := copy(p, d.data)

	if d.finished != nil {
		close(d.finished)
	}

	return count, d.err
}

const nextData = "NEXT"

// gateThenNextReader sends its first Read through gate and answers every
// later Read with nextData immediately.
type gateThenNextReader struct {
	gate   *gatedReader
	called atomic.Bool
}

func (r *gateThenNextReader) Read(p []byte) (int, error) {
	if r.called.CompareAndSwap(false, true) {
		return r.gate.Read(p)
	}

	return copy(p, nextData), nil
}
