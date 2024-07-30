package ioutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeWriter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var sb strings.Builder
		sw, err := NewSafeWriter(&sb)
		require.Nil(t, err)
		_, err = sw.Write([]byte("test"))
		require.Nil(t, err)
		require.Equal(t, "test", sb.String())
	})

	t.Run("failure", func(t *testing.T) {
		sw, err := NewSafeWriter(nil)
		require.NotNil(t, err)
		require.Nil(t, sw)
	})
}
