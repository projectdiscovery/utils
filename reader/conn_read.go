package reader

import (
	"bytes"
	"context"
	"errors"
	"io"
	"syscall"
	"time"

	contextutil "github.com/projectdiscovery/utils/context"
	errorutil "github.com/projectdiscovery/utils/errors"
)

const (
	// although this is more than enough for most cases
	MaxReadSize = 1 << 23 // 8MB
)

var (
	ErrTooLarge = errors.New("reader: too large only 8MB allowed as per MaxReadSize")
)

// ConnReadN reads at most N bytes from reader and it optimized
// for connection based readers like net.Conn it should not be used
// for file/buffer based reading, ConnReadN should be preferred
// instead of 'conn.Read() without loop'
// Note: you are responsible for adding a timeout to context
func ConnReadN(ctx context.Context, reader io.Reader, N int64) ([]byte, error) {
	if N == -1 {
		N = MaxReadSize
	} else if N < -1 {
		return nil, errors.New("reader: N cannot be less than -1")
	} else if N == 0 {
		return []byte{}, nil
	} else if N > MaxReadSize {
		return nil, ErrTooLarge
	}
	var buff bytes.Buffer
	fn := func() (int64, error) {
		return io.CopyN(&buff, io.LimitReader(reader, N), N)
	}
	_, err := contextutil.ExecFuncWithTwoReturns(ctx, fn)
	if err != nil {
		if IsAcceptedError(err) && buff.Len() > 0 {
			// if error is accepted error and we have some data
			// then return data
			return buff.Bytes(), nil
		} else {
			return nil, errorutil.WrapfWithNil(err, "reader: error while reading from connection")
		}
	} else {
		return buff.Bytes(), nil
	}
}

// ConnReadNWithTimeout is same as ConnReadN but it takes timeout
// instead of context and it returns error if read does not finish in given time
func ConnReadNWithTimeout(reader io.Reader, N int64, after time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), after)
	defer cancel()
	return ConnReadN(ctx, reader, N)
}

// IsAcceptedError checks if the error is accepted error
// for example: timeout, connection refused, io.EOF, io.ErrUnexpectedEOF
// while reading from connection
func IsAcceptedError(err error) bool {
	if err == io.EOF || err == io.ErrUnexpectedEOF || errorutil.IsTimeout(err) {
		// ideally we should error out if we get a timeout error but
		// that's different for our use case
		return true
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	return false
}
