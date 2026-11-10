//go:build darwin

package rawmode

import "syscall"

func init() {
	TCSETS = syscall.TIOCSETA
	TCGETS = syscall.TIOCGETA
}
