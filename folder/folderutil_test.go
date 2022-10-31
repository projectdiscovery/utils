package folderutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFiles(t *testing.T) {
	// get files from current folder
	files, err := GetFiles(".")
	require.Nilf(t, err, "couldn't retrieve the list of files: %s", err)

	// we check only if the number of files is bigger than zero
	require.Positive(t, len(files), "no files could be retrieved: %s", err)
}
