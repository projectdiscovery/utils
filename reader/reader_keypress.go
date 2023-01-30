package reader

import (
	"context"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/projectdiscovery/utils/reader/rawmode"
)

type KeyPressReader struct {
	mode     uint32
	Timeout  time.Duration
	datachan chan []byte
	Once     *sync.Once
}

func (reader *KeyPressReader) Start() error {
	reader.Once.Do(func() {
		go reader.read()
		reader.mode, _ = rawmode.GetMode(os.Stdin)
	})
	return rawmode.SetRawMode(os.Stdin, reader.mode)
}

func (reader *KeyPressReader) Stop() error {
	return rawmode.SetMode(os.Stdin, reader.mode)
}

func (reader *KeyPressReader) read() {
	if reader.datachan == nil {
		reader.datachan = make(chan []byte)
	}
	for {
		r := make([]byte, 1)
		n, err := syscall.Read(syscall.Handle(os.Stdin.Fd()), r)
		if n > 0 && err == nil {
			reader.datachan <- r
		}
	}
}

// Read into the buffer
func (reader KeyPressReader) Read(p []byte) (n int, err error) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if reader.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(reader.Timeout))
		defer cancel()
	}

	select {
	case <-ctx.Done():
		err = ErrTimeout
		return
	case data := <-reader.datachan:
		n = copy(p, data)
		return
	}
}
