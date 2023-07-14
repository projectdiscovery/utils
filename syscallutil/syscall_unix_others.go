//go:build (darwin || linux) && (386 || arm)

package syscallutil

import "github.com/ebitengine/purego"

func loadLibrary(name string) (uintptr, error) {
	return -1, errors.New("not implemented")
}
