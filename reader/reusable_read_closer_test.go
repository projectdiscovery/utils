package reader

import (
	"bytes"
	"io"
	"strings"
	"sync"
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
		t.Run("sequential reuse", func(t *testing.T) {
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
		})

		// todo: readers shouldn't be used concurrently, so here we just try to catch potential read concurring with resets panics
		t.Run("concurrent-reset-with-read", func(t *testing.T) {
			reusableReader, err := NewReusableReadCloser(v)
			require.Nil(t, err)

			var wg sync.WaitGroup

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					_, err := io.Copy(io.Discard, reusableReader)
					require.Nil(t, err)
					_, err = io.ReadAll(reusableReader)
					require.Nil(t, err)
				}()
			}

			wg.Wait()
		})
	}
}

// Readers must own their bytes. reset writes the buffer back at EOF, so a
// caller that reuses the slice, and a second reader built from it, must still
// see the original body.
func TestReusableReaderCopiesByteSlices(t *testing.T) {
	for name, build := range map[string]func([]byte) (*ReusableReadCloser, error){
		"[]byte":  func(b []byte) (*ReusableReadCloser, error) { return NewReusableReadCloser(b) },
		"*[]byte": func(b []byte) (*ReusableReadCloser, error) { return NewReusableReadCloser(&b) },
	} {
		t.Run(name, func(t *testing.T) {
			body := []byte("original")
			reader, err := build(body)
			require.NoError(t, err)
			copy(body, "modified")

			got, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.Equal(t, "original", string(got))
			require.Equal(t, "modified", string(body))
		})

		t.Run(name+" two readers", func(t *testing.T) {
			body := []byte("original")
			first, err := build(body)
			require.NoError(t, err)
			second, err := build(body)
			require.NoError(t, err)
			copy(body, "modified")

			for _, reader := range []*ReusableReadCloser{first, second, first} {
				got, err := io.ReadAll(reader)
				require.NoError(t, err)
				require.Equal(t, "original", string(got))
			}
			require.Equal(t, "modified", string(body))
		})
	}
}
