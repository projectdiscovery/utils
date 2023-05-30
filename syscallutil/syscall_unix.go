//go:build darwin || linux

package syscallutil

import "github.com/ebitengine/purego"

func loadLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
