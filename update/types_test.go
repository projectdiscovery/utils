package updateutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsOutdated(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{
			current:  "1.0.0",
			latest:   "1.1.0",
			expected: true,
		},
		{
			current:  "1.0.0",
			latest:   "1.0.0",
			expected: false,
		},
		{
			current:  "1.1.0",
			latest:   "1.0.0",
			expected: false,
		},
		{
			current:  "1.0.0-dev",
			latest:   "1.0.0",
			expected: true,
		},
		{
			current:  "invalid",
			latest:   "1.0.0",
			expected: true,
		},
		{
			current:  "invalid1",
			latest:   "invalid2",
			expected: true,
		},
		{
			current:  "1.0.0-alpha",
			latest:   "1.0.0",
			expected: true,
		},
		{
			current:  "1.0.0-alpha",
			latest:   "1.0.0-beta",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("current: %v, latest: %v", tt.current, tt.latest), func(t *testing.T) {
			assert.Equal(t, tt.expected, IsOutdated(tt.current, tt.latest), "version comparison failed")
		})
	}
}
