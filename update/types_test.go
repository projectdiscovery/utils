package updateutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsOutdated(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected bool
	}{
		{
			name:     "Current version is older than latest",
			current:  "1.0.0",
			latest:   "1.1.0",
			expected: true,
		},
		{
			name:     "Current version is same as latest",
			current:  "1.0.0",
			latest:   "1.0.0",
			expected: false,
		},
		{
			name:     "Current version is newer than latest",
			current:  "1.1.0",
			latest:   "1.0.0",
			expected: false,
		},
		{
			name:     "Current version is dev version",
			current:  "1.0.0-dev",
			latest:   "1.0.0",
			expected: true,
		},
		{
			name:     "Invalid version format - fallback to string comparison",
			current:  "invalid",
			latest:   "1.0.0",
			expected: true,
		},
		{
			name:     "Both versions invalid - fallback to string comparison",
			current:  "invalid1",
			latest:   "invalid2",
			expected: true,
		},
		{
			name:     "Pre-release version comparison",
			current:  "1.0.0-alpha",
			latest:   "1.0.0",
			expected: true,
		},
		{
			name:     "Pre-release version comparison with same base version",
			current:  "1.0.0-alpha",
			latest:   "1.0.0-beta",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsOutdated(tt.current, tt.latest), "version comparison failed")
		})
	}
}
