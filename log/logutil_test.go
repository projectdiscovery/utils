package logutil

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisableDefaultLogger(t *testing.T) {
	msg := "sample test"
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	DisableDefaultLogger()
	log.Print(msg)
	require.Equal(t, "", buf.String())
}

func TestEnableDefaultLogger(t *testing.T) {
	msg := "sample test"
	buf := new(bytes.Buffer)
	var stderr = *os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	exit := make(chan bool)
	go func() {
		_, _ = io.Copy(buf, r)
		exit <- true
	}()
	EnableDefaultLogger()
	log.Print(msg)
	w.Close()
	<-exit
	os.Stderr = &stderr
	require.Contains(t, buf.String(), msg)
}
