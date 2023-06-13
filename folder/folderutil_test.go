package folderutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFiles(t *testing.T) {
	// get files from current folder
	files, err := GetFiles(".")
	require.Nilf(t, err, "couldn't retrieve the list of files: %s", err)

	// we check only if the number of files is bigger than zero
	require.Positive(t, len(files), "no files could be retrieved: %s", err)
}

func TestMigrateDir(t *testing.T) {
	t.Run("destination folder creation error", func(t *testing.T) {
		err := MigrateDir("/source", "/:/dest")
		assert.Error(t, err)
	})

	t.Run("source folder not found error", func(t *testing.T) {
		err := MigrateDir("/notExistingFolder", "/dest")
		assert.Error(t, err)
	})

	t.Run("successful migration", func(t *testing.T) {
		// setup
		// some files in a temp dir
		sourceDir := t.TempDir()
		defer os.RemoveAll(sourceDir)
		_ = os.WriteFile(sourceDir+"/file1.txt", []byte("file1"), 0644)
		_ = os.WriteFile(sourceDir+"/file2.txt", []byte("file2"), 0644)

		// destination directory
		destinationDir := t.TempDir()
		defer os.RemoveAll(destinationDir)

		// when: try to migrate files
		err := MigrateDir(sourceDir, destinationDir)

		// then: verify if files migrated successfully
		assert.NoError(t, err)

		_, err = os.Stat(destinationDir + "/file1.txt")
		assert.NoError(t, err)
		_, err = os.Stat(destinationDir + "/file2.txt")
		assert.NoError(t, err)
		_, err = os.Stat(sourceDir)
		assert.Error(t, err)
	})
}
