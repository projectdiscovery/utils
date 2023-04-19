package updateutils

import (
	"testing"

	"github.com/logrusorgru/aurora"
)

func TestGetVersionDescription(t *testing.T) {
	Aurora = aurora.NewAurora(false)
	tests := []struct {
		current string
		latest  string
		want    string
	}{
		{
			current: "v2.9.1-dev",
			latest:  "v2.9.1",
			want:    "(outdated)",
		},
		{
			current: "v2.9.1-dev",
			latest:  "v2.9.2",
			want:    "(outdated)",
		},
		{
			current: "v2.9.1-dev",
			latest:  "v2.9.0",
			want:    "(development)",
		},
		{
			current: "v2.9.1",
			latest:  "v2.9.1",
			want:    "(latest)",
		},
		{
			current: "v2.9.1",
			latest:  "v2.9.2",
			want:    "(outdated)",
		},
	}
	for _, test := range tests {
		if GetVersionDescription(test.current, test.latest) != test.want {
			t.Errorf("GetVersionDescription(%v, %v) = %v, want %v", test.current, test.latest, GetVersionDescription(test.current, test.latest), test.want)
		}
	}
}
