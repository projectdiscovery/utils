package reader

import (
	"bytes"
	"crypto/tls"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	t.Run("Read From Connection", func(t *testing.T) {
		conn, err := tls.Dial("tcp", "projectdiscovery.io:443", &tls.Config{InsecureSkipVerify: true})
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		require.Nil(t, err, "could not connect to projectdiscovery.io over tls")
		defer conn.Close()
		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: projectdiscovery.io\r\nConnection: close\r\n\r\n"))
		require.Nil(t, err, "could not write to connection")
		data, err := ConnReadN(conn, -1)
		require.Nilf(t, err, "could not read from connection: %s", err)
		require.NotEmpty(t, data, "could not read from connection")
	})
}
