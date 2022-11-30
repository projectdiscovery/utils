package fileutil

import (
	"bytes"
	"io"
)

type ReusableReader struct {
	io.Reader
	readBuf *bytes.Buffer
	backBuf *bytes.Buffer
}

func NewReusableReader(r io.Reader) (io.ReadCloser, error) {
	readBuf := bytes.Buffer{}
	if _, err := readBuf.ReadFrom(r); err != nil {
		return nil, err
	}
	backBuf := bytes.Buffer{}
	reusableReader := ReusableReader{
		io.TeeReader(&readBuf, &backBuf),
		&readBuf,
		&backBuf,
	}

	return reusableReader, nil
}

func (r ReusableReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		r.reset()
	}
	return n, err
}

func (r ReusableReader) reset() {
	_, _ = io.Copy(r.readBuf, r.backBuf)
}

func (r ReusableReader) Close() error {
	return nil
}
