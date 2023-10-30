package reader

import (
	"bytes"
	"strings"
	"testing"
)

func TestConnReadN(t *testing.T) {
	t.Run("Test with N as -1", func(t *testing.T) {
		reader := strings.NewReader("Hello, World!")
		data, err := ConnReadN(reader, -1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if string(data) != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got '%s'", string(data))
		}
	})

	t.Run("Test with N as 0", func(t *testing.T) {
		reader := strings.NewReader("Hello, World!")
		data, err := ConnReadN(reader, 0)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(data) != 0 {
			t.Errorf("Expected empty, got '%s'", string(data))
		}
	})

	t.Run("Test with N greater than MaxReadSize", func(t *testing.T) {
		reader := bytes.NewReader(make([]byte, MaxReadSize+1))
		_, err := ConnReadN(reader, MaxReadSize+1)
		if err != ErrTooLarge {
			t.Errorf("Expected 'ErrTooLarge', got '%v'", err)
		}
	})

	t.Run("Test with N less than MaxReadSize", func(t *testing.T) {
		reader := strings.NewReader("Hello, World!")
		data, err := ConnReadN(reader, 5)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if string(data) != "Hello" {
			t.Errorf("Expected 'Hello', got '%s'", string(data))
		}
	})
}
