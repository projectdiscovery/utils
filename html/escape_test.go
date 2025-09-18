// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"testing"
)

func TestEscapeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic HTML characters",
			input:    `<>&"'`,
			expected: `&lt;&gt;&amp;&quot;&apos;`,
		},
		{
			name:     "extended characters with accents",
			input:    "café résumé naïve",
			expected: "caf&eacute; r&eacute;sum&eacute; na&iuml;ve",
		},
		{
			name:     "mathematical symbols",
			input:    "α + β = γ",
			expected: "&alpha; &plus; &beta; &equals; &gamma;",
		},
		{
			name:     "mixed content",
			input:    "Price: €50 & £30",
			expected: "Price&colon; &euro;50 &amp; &pound;30",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no special characters",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "copyright and trademark",
			input:    "© 2023 Company™",
			expected: "&copy; 2023 Company&trade;",
		},
		{
			name:     "two-character entities",
			input:    "≂̸ ≧̸",
			expected: "&nesim; &ngE;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeString(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeStringStd(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic HTML characters",
			input:    `<>&"'`,
			expected: `&lt;&gt;&amp;&#34;&#39;`,
		},
		{
			name:     "extended characters with accents",
			input:    "café résumé naïve",
			expected: "café résumé naïve",
		},
		{
			name:     "mathematical symbols",
			input:    "α + β = γ",
			expected: "α + β = γ",
		},
		{
			name:     "mixed content",
			input:    "Price: €50 & £30",
			expected: "Price: €50 &amp; £30",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no special characters",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "copyright and trademark",
			input:    "© 2023 Company™",
			expected: "© 2023 Company™",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeStringStd(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeStringStd(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUnescapeEscape(t *testing.T) {
	ss := []string{
		``,
		`abc def`,
		`café résumé`,
		`α + β = γ`,
		`Price: €50 & £30`,
		`© 2023 Company™`,
		`"<&>"αβγ`,
		`The special characters are: <, >, &, ', " and more like ©, ™, €`,
	}
	for _, s := range ss {
		escaped := EscapeString(s)
		unescaped := UnescapeString(escaped)
		if unescaped != s {
			t.Errorf("UnescapeString(EscapeString(%q)) = %q, want %q", s, unescaped, s)
		}
	}
}
