package updateutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownloadNucleiRelease(t *testing.T) {
	HideProgressBar = true
	gh, err := NewghReleaseDownloader("nuclei")
	require.Nil(t, err)
	_, err = gh.GetExecutableFromAsset()
	require.Nil(t, err)
}

func TestDownloadNucleiTemplate(t *testing.T) {
	HideProgressBar = true
}
