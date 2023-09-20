package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"

	semver "github.com/Masterminds/semver/v3"
)

func bumpVersion(fileName, varName, part string) (string, string, error) {
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		return "", "", fmt.Errorf("unable to get absolute path: %v", err)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return "", "", fmt.Errorf("could not parse file: %w", err)
	}

	var oldVersion, newVersion string

	ast.Inspect(node, func(n ast.Node) bool {
		if v, ok := n.(*ast.GenDecl); ok {
			for _, spec := range v.Specs {
				if s, ok := spec.(*ast.ValueSpec); ok {
					for idx, id := range s.Names {
						if id.Name == varName {
							oldVersion, _ = strconv.Unquote(s.Values[idx].(*ast.BasicLit).Value)
							v, err := semver.NewVersion(oldVersion)
							if err != nil || v.String() == "" {
								return false
							}
							var vInc func() semver.Version
							switch part {
							case "major":
								vInc = v.IncMajor
							case "minor":
								vInc = v.IncMinor
							case "", "patch":
								vInc = v.IncPatch
							default:
								return false
							}
							newVersion = "v" + vInc().String()
							s.Values[idx].(*ast.BasicLit).Value = fmt.Sprintf("`%s`", newVersion)
							return false
						}
					}
				}
			}
		}
		return true
	})

	if newVersion == "" {
		return oldVersion, newVersion, fmt.Errorf("failed to update the version")
	}

	f, err := os.OpenFile(fileName, os.O_RDWR, 0666)
	if err != nil {
		return oldVersion, newVersion, fmt.Errorf("could not open file: %w", err)
	}
	defer f.Close()

	if err := format.Node(f, fset, node); err != nil {
		return oldVersion, newVersion, fmt.Errorf("could not write to file: %w", err)
	}

	return oldVersion, newVersion, nil
}

func main() {
	var (
		fileName string
		varName  string
		part     string
	)

	flag.StringVar(&fileName, "file", "", "Go source file to parse")
	flag.StringVar(&varName, "var", "", "Variable to update")
	flag.StringVar(&part, "part", "patch", "Version part to increment (major, minor, patch)")

	flag.Parse()

	if fileName == "" || varName == "" {
		fmt.Println("Error: Both -file and -var are required")
		os.Exit(1)
	}
	oldVersion, newVersion, err := bumpVersion(fileName, varName, part)
	if err != nil {
		fmt.Printf("Error bumping version: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Bump from %s to %s\n", oldVersion, newVersion)
}
