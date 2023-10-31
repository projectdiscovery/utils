package reader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"syscall"
	"time"

	contextutil "github.com/projectdiscovery/utils/context"
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
	fn := func() ([]byte, error) {
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
		// read N bytes or until EOF
		_, err := io.CopyN(&buff, io.LimitReader(reader, N), N)
		if err != nil && !IsAcceptedError(err) {
			return nil, err
		}
		return buff.Bytes(), nil
	}
	return contextutil.ExecFuncWithTwoReturns(ctx, fn)
}

// ConnReadNWithTimeout is same as ConnReadN but it takes timeout
// instead of context and it returns error if read does not finish in given time
func ConnReadNWithTimeout(reader io.Reader, N int64, after time.Duration) ([]byte, error) {
	ctx, cancel := context.WithDeadlineCause(context.Background(), time.Now().Add(after), fmt.Errorf("conn read deadline exceeded"))
	defer cancel()
	return ConnReadN(ctx, reader, N)
}

// IsAcceptedError checks if the error is accepted error
// for example: timeout, connection refused, io.EOF, io.ErrUnexpectedEOF
// while reading from connection
func IsAcceptedError(err error) bool {
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return true
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	return false
}
