package ioutil

import (
	"io"
	"sync"
)

type SafeWriter struct {
	writer io.Writer
	mutex  sync.Mutex
}

func NewSafeWriter(writer io.Writer) *SafeWriter {
	return &SafeWriter{
		writer: writer,
	}
}

func (sw *SafeWriter) Write(p []byte) (n int, err error) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	if sw.writer == nil {
		return 0, io.ErrClosedPipe
	}
	return sw.writer.Write(p)
}
