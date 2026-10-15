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

func TestReusableReader_DoesNotShareCallerSlice(t *testing.T) {
	orig := []byte("original payload")
	r, err := NewReusableReadCloser(orig)
	require.NoError(t, err)

	// Read all and verify
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "original payload", string(data))

	// Mutate original slice - reader's next read should not be affected
	orig[0] = 'X'
	data2, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, "original payload", string(data2))

	// Ensure caller slice was not overwritten by reset
	require.Equal(t, byte('X'), orig[0])

	// Same check with *[]byte
	slicePtr := &orig
	rPtr, err := NewReusableReadCloser(slicePtr)
	require.NoError(t, err)
	orig[0] = 'Y'
	pData, err := io.ReadAll(rPtr)
	require.NoError(t, err)
	require.Equal(t, "Xriginal payload", string(pData))

	// Same check with *bytes.Buffer
	buf := bytes.NewBuffer([]byte("buffer payload"))
	rBuf, err := NewReusableReadCloser(buf)
	require.NoError(t, err)
	buf.Reset()
	bData, err := io.ReadAll(rBuf)
	require.NoError(t, err)
	require.Equal(t, "buffer payload", string(bData))
}

func TestReusableReader_MultipleReadersFromSameSlice(t *testing.T) {
	shared := []byte("shared memory stream")
	r1, err := NewReusableReadCloser(shared)
	require.NoError(t, err)
	r2, err := NewReusableReadCloser(shared)
	require.NoError(t, err)

	// Concurrent / interleaved reads from readers built from the same slice
	b1, err := io.ReadAll(r1)
	require.NoError(t, err)
	require.Equal(t, "shared memory stream", string(b1))

	b2, err := io.ReadAll(r2)
	require.NoError(t, err)
	require.Equal(t, "shared memory stream", string(b2))

	// Second round of reads after reset
	b1Second, err := io.ReadAll(r1)
	require.NoError(t, err)
	require.Equal(t, "shared memory stream", string(b1Second))

	b2Second, err := io.ReadAll(r2)
	require.NoError(t, err)
	require.Equal(t, "shared memory stream", string(b2Second))
}

func TestReusableReader_NilPointers(t *testing.T) {
	var nilSlice *[]byte
	r1, err := NewReusableReadCloser(nilSlice)
	require.NoError(t, err)
	data, err := io.ReadAll(r1)
	require.NoError(t, err)
	require.Empty(t, data)

	var nilBuf *bytes.Buffer
	r2, err := NewReusableReadCloser(nilBuf)
	require.NoError(t, err)
	data2, err := io.ReadAll(r2)
	require.NoError(t, err)
	require.Empty(t, data2)
}

