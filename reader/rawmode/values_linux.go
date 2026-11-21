//go:build linux

package rawmode

import "syscall"

func init() {
	TCSETS = syscall.TCSETS
	TCGETS = syscall.TCGETS
}
