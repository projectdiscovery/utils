package reader

import (
	"io"
	"time"
)

// TimeoutReader is a reader wrapper that stops waiting after Timeout.
//
// A Timeout of zero or less disables the timeout: Read calls Reader directly.
//
// io.Reader has no way to cancel a Read that is blocked, so a timeout only
// stops the caller from waiting. The underlying Read keeps running on a helper
// goroutine until Reader returns, so a Reader that never returns leaks that
// goroutine. On timeout, Read returns 0 and ErrTimeout and never writes to the
// caller's buffer after it returns.
//
// Bytes (and the error) that Reader delivers after a timeout are discarded,
// not kept for the next Read, so the data stream has a gap. A Read after a
// timeout starts a new Read on Reader while the abandoned one may still be
// pending, so Reader must tolerate overlapping Reads. Treat ErrTimeout as
// terminal unless the data in flight does not matter.
//
// Every Read with a positive Timeout allocates a scratch buffer as long as the
// caller's, because Reader must not be handed the caller's buffer (see above).
type TimeoutReader struct {
	Timeout time.Duration
	Reader  io.Reader
}

// Read into the buffer
func (reader TimeoutReader) Read(buf []byte) (int, error) {
	if reader.Timeout <= 0 {
		return reader.Reader.Read(buf)
	}

	type result struct {
		n   int
		err error
	}

	// Reader gets a private buffer: after a timeout it may still write to it,
	// but buf is the caller's again. done has room for the late result so the
	// helper goroutine exits as soon as Reader returns, with nobody receiving.
	scratch := make([]byte, len(buf))
	done := make(chan result, 1)
	go func() {
		n, err := reader.Reader.Read(scratch)
		done <- result{n, err}
	}()

	timer := time.NewTimer(reader.Timeout)
	defer timer.Stop()

	select {
	case res := <-done:
		return copy(buf, scratch[:res.n]), res.err
	case <-timer.C:
		return 0, ErrTimeout
	}
}
