package logutils

import (
	"io"
	"log"
)

func init() {
	// disable standard logger (ref: https://github.com/golang/go/issues/19895)
	log.SetFlags(0)
	log.SetOutput(io.Discard)
}
