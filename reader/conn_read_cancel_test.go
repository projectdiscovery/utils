package reader

import (
	"context"
	"io"
	"net"
	"runtime"
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
	addr := newSilentServer(t)

	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	runtime.GC()
	base := runtime.NumGoroutine()

	const calls = 50
	for range calls {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		_, _ = ConnReadN(ctx, conn, 16) // peer sends nothing; ctx expires first
		cancel()
		t.Cleanup(func() { _ = conn.Close() })
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		runtime.GC()
		leaked := runtime.NumGoroutine() - base
		if leaked <= 2 { // tolerance for transient runtime goroutines
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("goroutine leak: ~%d goroutines still alive after %d context-cancelled ConnReadN calls (base=%d, now=%d)",
				leaked, calls, base, runtime.NumGoroutine())
		}
		time.Sleep(25 * time.Millisecond)
	}
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
		defer c.Close()
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

// newSilentServer returns the address of a TCP server that accepts connections
// and holds them open without ever writing, so reads against it block until the
// reader's deadline or close.
func newSilentServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	held := make(chan net.Conn, 1024)
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			held <- c // hold the server end open; never write
		}
	}()
	t.Cleanup(func() { _ = ln.Close() })
	return ln.Addr().String()
}
