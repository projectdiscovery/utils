package reader

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReusableReader(t *testing.T) {
	testcases := []interface{}{
		strings.NewReader("test"),
		bytes.NewBuffer([]byte("test")),
		bytes.NewBufferString("test"),
		bytes.NewReader([]byte("test")),
		[]byte("test"),
		"test",
	}
	for _, v := range testcases {
		reusableReader, err := NewReusableReadCloser(v)
		require.Nil(t, err)

		for i := 0; i < 100; i++ {
			n, err := io.Copy(io.Discard, reusableReader)
			require.Nil(t, err)
			require.Positive(t, n)

			bin, err := io.ReadAll(reusableReader)
			require.Nil(t, err)
			require.Len(t, bin, 4)
		}
	}
}
