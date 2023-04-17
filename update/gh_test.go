//go:build update

// update related tests are only executed when update tag is provided (ex: go test -tags update ./...) to avoid failures due to rate limiting
package updateutils

import (
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDownloadNucleiRelease tests downloading nuclei release
func TestDownloadNucleiRelease(t *testing.T) {
	HideProgressBar = true
	gh, err := NewghReleaseDownloader("nuclei")
	require.Nil(t, err)
	_, err = gh.GetExecutableFromAsset()
	require.Nil(t, err)
}

// TestDownloadNucleiTemplatesFromSource tests downloading nuclei-templates from source
func TestDownloadNucleiTemplatesFromSource(t *testing.T) {
	gh, err := NewghReleaseDownloader("nuclei-templates")
	require.Nil(t, err)
	counter := 0
	callback := func(path string, fileInfo fs.FileInfo, data io.Reader) error {
		_ = fileInfo.Name()
		counter++
		return nil
	}
	err = gh.DownloadSourceWithCallback(false, callback)
	require.Nil(t, err)
	// actual content is lot more than 100 files
	require.Greater(t, counter, 100)
}

// TestDownloadToolWithDifferentName tests downloading a tool with different name than repo name
// by default repo name is considered as executable name
func TestDownloadToolWithDifferentName(t *testing.T) {
	gh, err := NewghReleaseDownloader("interactsh")
	require.Nil(t, err)
	gh.SetToolName("interactsh-client")
	_, err = gh.GetExecutableFromAsset()
	require.Nil(t, err)
}
