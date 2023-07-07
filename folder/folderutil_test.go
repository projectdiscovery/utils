package folderutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	fileutil "github.com/projectdiscovery/utils/file"
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

	t.Run("source and destination are the same", func(t *testing.T) {
		// setup
		// some files in a temp dir
		sourceDir := t.TempDir()
		defer os.RemoveAll(sourceDir)
		_ = os.WriteFile(filepath.Join(sourceDir, "/file1.txt"), []byte("file1"), os.ModePerm)
		_ = os.WriteFile(filepath.Join(sourceDir, "/file2.txt"), []byte("file2"), os.ModePerm)

		// when: try to migrate files
		err := MigrateDir(sourceDir, sourceDir)

		// then: verify if files migrated successfully
		assert.Error(t, err)

		assert.True(t, fileutil.FileExists(filepath.Join(sourceDir, "/file1.txt")))
		assert.True(t, fileutil.FileExists(filepath.Join(sourceDir, "/file2.txt")))
	})

	t.Run("successful migration with source dir removal", func(t *testing.T) {
		// setup
		// some files in a temp dir
		sourceDir := t.TempDir()
		defer os.RemoveAll(sourceDir)
		_ = os.WriteFile(filepath.Join(sourceDir, "/file1.txt"), []byte("file1"), os.ModePerm)
		_ = os.WriteFile(filepath.Join(sourceDir, "/file2.txt"), []byte("file2"), os.ModePerm)
		_ = os.Mkdir(filepath.Join(sourceDir, "/dir1"), os.ModePerm)
		_ = os.WriteFile(filepath.Join(sourceDir, "/dir1", "/file3.txt"), []byte("file3"), os.ModePerm)
		_ = os.Mkdir(filepath.Join(sourceDir, "/dir2"), os.ModePerm)

		// destination directory
		destinationDir := t.TempDir()
		defer os.RemoveAll(destinationDir)

		// when: try to migrate files
		err := MigrateDir(sourceDir, destinationDir)

		// then: verify if files migrated successfully
		assert.NoError(t, err, sourceDir, destinationDir)

		assert.True(t, fileutil.FileExists(filepath.Join(destinationDir, "/file1.txt")))
		assert.True(t, fileutil.FileExists(filepath.Join(destinationDir, "/file2.txt")))
		assert.True(t, fileutil.FolderExists(filepath.Join(destinationDir, "/dir1")))
		assert.True(t, fileutil.FileExists(filepath.Join(destinationDir, "/dir1", "/file3.txt")))

		assert.False(t, fileutil.FolderExists(filepath.Join(destinationDir, "/dir2")))

		assert.False(t, fileutil.FolderExists(sourceDir))
	})

	t.Run("successful migration without source dir removal", func(t *testing.T) {
		// setup
		// some files in a temp dir
		sourceDir := t.TempDir()
		defer os.RemoveAll(sourceDir)
		_ = os.WriteFile(filepath.Join(sourceDir, "/file1.txt"), []byte("file1"), os.ModePerm)
		_ = os.WriteFile(filepath.Join(sourceDir, "/file2.txt"), []byte("file2"), os.ModePerm)
		_ = os.Mkdir(filepath.Join(sourceDir, "/dir1"), os.ModePerm)
		_ = os.WriteFile(filepath.Join(sourceDir, "/dir1", "/file3.txt"), []byte("file3"), os.ModePerm)
		_ = os.Mkdir(filepath.Join(sourceDir, "/dir2"), os.ModePerm)

		// destination directory
		destinationDir := t.TempDir()
		defer os.RemoveAll(destinationDir)

		// when: try to migrate files
		RemoveSourceDirAfterMigration = false
		err := MigrateDir(sourceDir, destinationDir)

		// then: verify if files migrated successfully
		assert.NoError(t, err)

		assert.True(t, fileutil.FileExists(filepath.Join(destinationDir, "/file1.txt")))
		assert.True(t, fileutil.FileExists(filepath.Join(destinationDir, "/file2.txt")))
		assert.True(t, fileutil.FolderExists(filepath.Join(destinationDir, "/dir1")))
		assert.True(t, fileutil.FileExists(filepath.Join(destinationDir, "/dir1", "/file3.txt")))

		assert.False(t, fileutil.FolderExists(filepath.Join(destinationDir, "/dir2")))

		assert.True(t, fileutil.FolderExists(sourceDir))
	})
}

func TestMustMigrateDir(t *testing.T) {
	t.Run("it should exit if MigrateDir returns an error", func(t *testing.T) {
		// // given
		sourceDir := "/notExistingFolder"
		destinationDir := "/dest"

		ExitsOnFailure(t, func() {
			MustMigrateDir(sourceDir, destinationDir)
		})
	})
}

func ExitsOnFailure(t *testing.T, f func()) {
	if os.Getenv("BE_CRASHER") == "1" {
		f()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v, want exit status 1", err)
}
