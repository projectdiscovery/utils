package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/projectdiscovery/utils/detach"
)

var targetFile = filepath.Join("/tmp", "detach.test.txt")

func main() {
	if err := detach.Run(detachedFunc); err != nil {
		log.Fatal(err)
	}
}

func detachedFunc() error {
	time.Sleep(10 * time.Second)

	_ = os.WriteFile(targetFile, []byte("test"), fs.ModePerm)

	return nil
}
