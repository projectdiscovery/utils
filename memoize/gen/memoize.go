package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	fileutil "github.com/projectdiscovery/utils/file"
	"github.com/projectdiscovery/utils/memoize"
)

var (
	srcFolder   = flag.String("src", "", "source folder")
	dstFolder   = flag.String("dst", "", "destination foldder")
	packageName = flag.String("pkg", "memo", "destination package")
)

func main() {
	flag.Parse()

	_ = fileutil.CreateFolder(*dstFolder)

	err := filepath.WalkDir(*srcFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); strings.ToLower(ext) != ".go" {
			return nil
		}

		return process(path)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func process(path string) error {
	filename := filepath.Base(path)
	dstFile := filepath.Join(*dstFolder, filename)
	out, err := memoize.File(path, *packageName)
	if err != nil {
		return err
	}
	return os.WriteFile(dstFile, out, os.ModePerm)
}
