package httputil

import (
	"bytes"
	"io"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseChain_ZstdLimitedBody(t *testing.T) {
	// Exercise concurrent decoding even on machines with a single CPU.
	previous := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previous)

	for _, tc := range []struct {
		name        string
		contentType string
		encoded     []byte
		decoded     []byte
		repeats     int
		limit       int
	}{
		{name: "plain", encoded: []byte("A"), decoded: []byte("A"), repeats: 4 << 20, limit: 32 << 10},
		{name: "gbk", contentType: "text/plain; charset=gbk", encoded: []byte{0xd6, 0xd0}, decoded: []byte("中"), repeats: 2 << 20, limit: 32 << 10},
		{name: "windows1251", contentType: "text/plain; charset=windows-1251", encoded: []byte{0xdf}, decoded: []byte("Я"), repeats: 4 << 20, limit: 32 << 10},
		{name: "10MB_limit", encoded: []byte("A"), decoded: []byte("A"), repeats: 16 << 20, limit: 10 << 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer leaktest.CheckTimeout(t, 2*time.Second)()
			compressed := encodeZstd(t, bytes.Repeat(tc.encoded, tc.repeats))
			require.Less(t, len(compressed), tc.limit)
			expected := bytes.Repeat(tc.decoded, tc.repeats)[:tc.limit]

			for range 3 {
				response := &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Encoding": []string{"zstd"},
						"Content-Type":     []string{tc.contentType},
					},
					Body: io.NopCloser(bytes.NewReader(compressed)),
				}
				chain := NewResponseChain(response, int64(tc.limit))
				err := chain.Fill()
				if err == nil {
					assert.Equal(t, expected, chain.BodyBytes())
				}
				chain.Close()
				require.NoError(t, err)
			}
		})
	}
}

func TestResponseChain_ZstdReadResults(t *testing.T) {
	previous := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(previous)

	original := bytes.Repeat([]byte("zstd response\n"), 1024)
	compressed := encodeZstd(t, original)
	for _, tc := range []struct {
		name    string
		data    []byte
		wantErr bool
		partial bool
	}{
		{name: "complete", data: compressed},
		{name: "invalid", data: []byte("not a zstd stream"), wantErr: true},
		{name: "truncated", data: bytes.Join([][]byte{compressed, compressed[:len(compressed)-1]}, nil), partial: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer leaktest.CheckTimeout(t, 2*time.Second)()
			response := &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Encoding": []string{"zstd"}},
				Body:       io.NopCloser(bytes.NewReader(tc.data)),
			}
			defer func() {
				require.NoError(t, response.Body.Close())
			}()
			chain := NewResponseChain(response, 32<<10)
			defer chain.Close()

			err := chain.Fill()
			if tc.wantErr {
				require.ErrorContains(t, err, "could not read response body")
				return
			}
			require.NoError(t, err)
			require.Equal(t, original, chain.BodyBytes())
			if tc.partial {
				require.Contains(t, response.Header.Get("x-nuclei-ignore-error"), "unexpected EOF")
			} else {
				require.Empty(t, response.Header.Get("x-nuclei-ignore-error"))
			}
		})
	}
}

func encodeZstd(tb testing.TB, data []byte) []byte {
	tb.Helper()
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
	require.NoError(tb, err)
	compressed := encoder.EncodeAll(data, nil)
	require.NoError(tb, encoder.Close())
	return compressed
}

func BenchmarkResponseChain_Zstd(b *testing.B) {
	for _, tc := range []struct {
		name  string
		size  int
		limit int64
	}{
		{name: "small", size: 1024, limit: 32 << 10},
		{name: "full", size: 4 << 20, limit: 8 << 20},
		{name: "limited", size: 4 << 20, limit: 32 << 10},
	} {
		b.Run(tc.name, func(b *testing.B) {
			original := bytes.Repeat([]byte("A"), tc.size)
			compressed := encodeZstd(b, original)

			b.ReportAllocs()
			b.SetBytes(min(int64(tc.size), tc.limit))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				response := &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Encoding": []string{"zstd"}},
					Body:       io.NopCloser(bytes.NewReader(compressed)),
				}
				chain := NewResponseChain(response, tc.limit)
				err := chain.Fill()
				chain.Close()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
