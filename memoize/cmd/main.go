package main

import (
	"log"

	"github.com/projectdiscovery/utils/memoize"
)

func main() {
	out, err := memoize.File(memoize.PackageTemplate, "../tests/test.go", "test")
	if err != nil {
		panic(err)
	}
	log.Println(string(out))
}
