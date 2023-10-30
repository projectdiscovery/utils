package reader

import (
	"bytes"
	"errors"
	"io"
)

const (
	// although this is more than enough for most cases
	MaxReadSize   = 1 << 23 // 8MB
	BuffAllocSize = 1 << 12 // 4KB
)

var (
	ErrTooLarge = errors.New("reader: too large only 8MB allowed as per MaxReadSize")
)

// ConnReadN reads at most N bytes from reader and it optimized
// for connection based readers like net.Conn it should not be used
// for file/buffer based reading, ConnReadN should be preferred
// instead of 'conn.Read() without loop'
func ConnReadN(reader io.Reader, N int64) ([]byte, error) {
	if N == -1 {
		N = MaxReadSize
	} else if N < -1 {
		return nil, errors.New("reader: N cannot be less than -1")
	} else if N == 0 {
		return []byte{}, nil
	} else if N > MaxReadSize {
		return nil, ErrTooLarge
	}
	// no need to allocate slice upfront if N is signficantly larger than BuffAllocSize
	// since we use this to read from network connection lets read 4KB at a time
	// net.Conn has 2KB and 4KB variants while reading http request from net.Conn
	allocSize := N
	if N > BuffAllocSize {
		allocSize = BuffAllocSize
	}

	buff := bytes.NewBuffer(make([]byte, 0, allocSize))
	// read N bytes or until EOF
	_, err := io.CopyN(buff, reader, N)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return buff.Bytes(), nil
}
