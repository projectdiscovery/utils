package permissionutil

import (
	"testing"

	osutils "github.com/projectdiscovery/utils/os"
	"github.com/stretchr/testify/require"
)

func TestIsRoot(t *testing.T) {
	t.Run("windows - linux", func(t *testing.T) {
		isRoot, err := checkCurrentUserRoot()
		if osutils.IsWindows() || osutils.IsLinux() {
			require.Nil(t, err)
			require.NotNil(t, isRoot)
		}
	})
}
