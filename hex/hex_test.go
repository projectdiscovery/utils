package hexutil

import (
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		format   []string
		expected string
	}{
		{
			name:     "string data standard format",
			data:     "meow",
			format:   nil,
			expected: "6d656f77",
		},
		{
			name:     "string data escaped format",
			data:     "meow",
			format:   []string{"x"},
			expected: "\\x6d\\x65\\x6f\\x77",
		},
		{
			name:     "byte slice data standard format",
			data:     []byte("meow"),
			format:   nil,
			expected: "6d656f77",
		},
		{
			name:     "byte slice data escaped format",
			data:     []byte("meow"),
			format:   []string{"x"},
			expected: "\\x6d\\x65\\x6f\\x77",
		},
		{
			name:     "empty string standard format",
			data:     "",
			format:   nil,
			expected: "",
		},
		{
			name:     "empty string escaped format",
			data:     "",
			format:   []string{"x"},
			expected: "",
		},
		{
			name:     "single character standard format",
			data:     "a",
			format:   nil,
			expected: "61",
		},
		{
			name:     "single character escaped format",
			data:     "a",
			format:   []string{"x"},
			expected: "\\x61",
		},
		{
			name:     "unknown format defaults to standard",
			data:     "meow",
			format:   []string{"unknown"},
			expected: "6d656f77",
		},
		{
			name:     "uppercase format",
			data:     "meow",
			format:   []string{"X"},
			expected: "\\x6d\\x65\\x6f\\x77",
		},
		{
			name:     "non-string data defaults to empty",
			data:     123,
			format:   nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Encode(tt.data, tt.format...)
			if result != tt.expected {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeStandard(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		expected string
	}{
		{
			name:     "string data",
			data:     "meow",
			expected: "6d656f77",
		},
		{
			name:     "byte slice data",
			data:     []byte("meow"),
			expected: "6d656f77",
		},
		{
			name:     "empty string",
			data:     "",
			expected: "",
		},
		{
			name:     "single character",
			data:     "a",
			expected: "61",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeStandard(tt.data)
			if result != tt.expected {
				t.Errorf("EncodeStandard() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeEscaped(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		expected string
	}{
		{
			name:     "string data",
			data:     "meow",
			expected: "\\x6d\\x65\\x6f\\x77",
		},
		{
			name:     "byte slice data",
			data:     []byte("meow"),
			expected: "\\x6d\\x65\\x6f\\x77",
		},
		{
			name:     "empty string",
			data:     "",
			expected: "",
		},
		{
			name:     "single character",
			data:     "a",
			expected: "\\x61",
		},
		{
			name:     "unicode string",
			data:     "hello 世界",
			expected: "\\x68\\x65\\x6c\\x6c\\x6f\\x20\\xe4\\xb8\\x96\\xe7\\x95\\x8c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeEscaped(tt.data)
			if result != tt.expected {
				t.Errorf("EncodeEscaped() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeEdgeCases(t *testing.T) {
	t.Run("odd length hex string", func(t *testing.T) {
		data := "a"
		result := EncodeEscaped(data)
		expected := "\\x61"
		if result != expected {
			t.Errorf("EncodeEscaped() = %v, want %v", result, expected)
		}
	})

	t.Run("very long string", func(t *testing.T) {
		data := strings.Repeat("a", 1000)
		result := EncodeEscaped(data)
		expectedLength := 1000 * 4
		if len(result) != expectedLength {
			t.Errorf("EncodeEscaped() length = %v, want %v", len(result), expectedLength)
		}
	})
}
