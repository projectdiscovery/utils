package httputil

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func BenchmarkResponseChain_LargeBody(b *testing.B) {
	body := bytes.Repeat([]byte("G"), 1024*1024) // 1MB

	b.Run("Body().Bytes()", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader(body)),
				Header:     http.Header{},
			}

			rc := NewResponseChain(resp, -1)
			_ = rc.Fill()
			_ = rc.Body().Bytes()
			rc.Close()
		}
	})
}

func BenchmarkResponseChain_StringConversion(b *testing.B) {
	body := bytes.Repeat([]byte("H"), 1024*1024) // 1MB
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     http.Header{},
	}

	rc := NewResponseChain(resp, -1)
	_ = rc.Fill()
	defer rc.Close()

	b.Run("Body().String()", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = rc.Body().String()
		}
	})

	b.Run("BodyString", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = rc.BodyString()
		}
	})
}
