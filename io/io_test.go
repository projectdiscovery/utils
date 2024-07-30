package ioutil

import (
	"strings"
	"testing"
)

func TestSafeWriter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var sb strings.Builder
		sw := NewSafeWriter(&sb)
		_, err := sw.Write([]byte("test"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if sb.String() != "test" {
			t.Fatalf("expected 'test', got '%s'", sb.String())
		}
	})

	t.Run("failure", func(t *testing.T) {
		sw := NewSafeWriter(nil)
		_, err := sw.Write([]byte("test"))
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
