package ioutil

import (
	"errors"
	"io"
	"sync"
)

// SafeWriter is a thread-safe wrapper for io.Writer
type SafeWriter struct {
	writer io.Writer   // The underlying writer
	mutex  *sync.Mutex // Mutex for ensuring thread-safety
}

// NewSafeWriter creates and returns a new SafeWriter
func NewSafeWriter(writer io.Writer) (*SafeWriter, error) {
	// Check if the provided writer is nil
	if writer == nil {
		return nil, errors.New("writer is nil")
	}

	safeWriter := &SafeWriter{
		writer: writer,
		mutex:  &sync.Mutex{},
	}
	return safeWriter, nil
}

// Write implements the io.Writer interface in a thread-safe manner
func (sw *SafeWriter) Write(p []byte) (n int, err error) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	if sw.writer == nil {
		return 0, io.ErrClosedPipe
	}
	return sw.writer.Write(p)
}
