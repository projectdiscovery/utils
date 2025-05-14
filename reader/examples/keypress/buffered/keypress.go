package main

import (
	"log"
	"sync"
	"time"

	"github.com/projectdiscovery/utils/reader"
	stringsutil "github.com/projectdiscovery/utils/strings"
)

func main() {
	stdr := reader.KeyPressReader{
		Timeout: time.Duration(5 * time.Second),
		Once:    &sync.Once{},
	}

	_ = stdr.Start()
	defer func() {
		_ = stdr.Stop()
	}()

	for {
		data := make([]byte, stdr.BufferSize)
		n, err := stdr.Read(data)
		log.Println(n, err)

		if stringsutil.IsCTRLC(string(data)) {
			break
		}
	}
}
