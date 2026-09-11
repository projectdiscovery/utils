package reader

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// TestConnReadN_NoGoroutineLeakOnCancel guards against ConnReadN leaking its
// read goroutine when the context is cancelled before the peer sends data.
//
// ConnReadN performs the read in a goroutine and returns the context error when
// ctx is done. A context deadline can't interrupt a blocking socket read, so
// unless ConnReadN actively unblocks the read on cancellation the goroutine
// stays parked in Read for the connection's lifetime, leaking the goroutine,
// its buffers, and the connection on every cancelled read.
func TestConnReadN_NoGoroutineLeakOnCancel(t *testing.T) {
	const calls = 50
	for range calls {
		reader := newDeadlineReader()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		_, _ = ConnReadN(ctx, reader, 16)
		cancel()

		select {
		case <-reader.readDone:
		case <-time.After(time.Second):
			reader.release()
			t.Fatal("read goroutine remained blocked after context cancellation")
		}
	}
}

type deadlineReader struct {
	unblock  chan struct{}
	readDone chan struct{}
	once     sync.Once
}

func newDeadlineReader() *deadlineReader {
	return &deadlineReader{
		unblock:  make(chan struct{}),
		readDone: make(chan struct{}),
	}
}

func (r *deadlineReader) Read(_ []byte) (int, error) {
	<-r.unblock
	close(r.readDone)
	return 0, context.DeadlineExceeded
}

func (r *deadlineReader) SetReadDeadline(_ time.Time) error {
	r.release()
	return nil
}

func (r *deadlineReader) release() {
	r.once.Do(func() { close(r.unblock) })
}

// TestConnReadN_ReturnsPartialDataOnCancel verifies that data already received
// before the context is cancelled is returned rather than dropped. Expiring the
// read deadline to unblock the read produces a net timeout error; ConnReadN must
// report the cancellation as the context error so the partial data is returned.
func TestConnReadN_ReturnsPartialDataOnCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = c.Close() }()
		_, _ = c.Write([]byte("hi"))  // send partial data, then stall
		_, _ = io.Copy(io.Discard, c) // block until the client closes
	}()

	client, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// asks for 16 bytes but only 2 arrive before the context expires
	data, err := ConnReadN(ctx, client, 16)
	if err != nil {
		t.Fatalf("expected partial data with no error, got err: %v", err)
	}
	if string(data) != "hi" {
		t.Fatalf("expected %q, got %q", "hi", string(data))
	}
}
