package fileutil

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReusableReader(t *testing.T) {
	reusableReader, err := NewReusableReader(strings.NewReader("test"))
	require.Nil(t, err)

	for i := 0; i < 100; i++ {
		n, err := io.Copy(io.Discard, reusableReader)
		require.Nil(t, err)
		require.Positive(t, n)
	}
}
