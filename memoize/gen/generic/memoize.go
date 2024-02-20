// this small cli tool is specific for those functions with arbitrary parameters and with result-error tuple as return values
// func(x,y) => result, error
// it works by creating a new memoized version of the functions in the same path as memo.original.file.go
// some parts are specific for nuclei and hardcoded within the template
package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/projectdiscovery/utils/memoize"
	stringsutil "github.com/projectdiscovery/utils/strings"
)

const (
	nucleiPath string = "/Users/marcornvh/go/src/github.com/projectdiscovery/nuclei/pkg/js/libs"
)

func main() {
	err := filepath.Walk(nucleiPath, walk)
	if err != nil {
		log.Fatal(err)
	}
}

func walk(path string, info fs.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}

	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	base := filepath.Base(path)

	if !stringsutil.EqualFoldAny(ext, ".go") {
		return nil
	}

	basePath := filepath.Dir(path)
	outPath := filepath.Join(basePath, "memo."+base)

	// filename := filepath.Base(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !stringsutil.ContainsAnyI(string(data), "@memo") {
		return nil
	}
	out, err := memoize.Src(memoize.PackageTemplate, path, data, "test")
	if err != nil {
		return err
	}

	if err := os.WriteFile(outPath, out, os.ModePerm); err != nil {
		return err
	}

	return nil
}
