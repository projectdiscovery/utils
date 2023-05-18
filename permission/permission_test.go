//go:build windows || linux

package permissionutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsRoot(t *testing.T) {
	isRoot, err := checkCurrentUserRoot()
	require.Nil(t, err)
	require.NotNil(t, isRoot)
}

func TestFilePermissions(t *testing.T) {

	const testFilePath = "/tmp/testfile.txt"

	t.Run("TestFileAllReadWriteExecute", func(t *testing.T) {
		file, err := os.Create(testFilePath)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFilePath)
		defer file.Close()

		// Set the file permissions
		err = file.Chmod(os.FileMode(AllReadWriteExecute))
		if err != nil {
			t.Fatalf("Failed to set file permissions: %v", err)
		}

		// Get the file permissions
		fileInfo, err := os.Stat(testFilePath)
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}
		// Check if the file permissions match the defined constants
		if fileInfo.Mode().Perm().String() != "-rwxrwxrwx" || fileInfo.Mode().Perm() != os.FileMode(AllReadWriteExecute) {
			t.Errorf("File permissions do not match. Expected: %s, Actual: %s", os.FileMode(AllReadWriteExecute).String(), fileInfo.Mode().Perm().String())
		}
	})

	t.Run("TestFileUserReadWriteExecute", func(t *testing.T) {
		file, err := os.Create(testFilePath)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFilePath)
		defer file.Close()

		// Set the file permissions
		err = file.Chmod(os.FileMode(UserReadWriteExecute))
		if err != nil {
			t.Fatalf("Failed to set file permissions: %v", err)
		}

		// Get the file permissions
		fileInfo, err := os.Stat(testFilePath)
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}
		// Check if the file permissions match the defined constants
		if fileInfo.Mode().Perm().String() != "-rwx------" || fileInfo.Mode().Perm() != os.FileMode(UserReadWriteExecute) {
			t.Errorf("File permissions do not match. Expected: %s, Actual: %s", os.FileMode(UserReadWriteExecute).String(), fileInfo.Mode().Perm().String())
		}
	})

	t.Run("TestFileGroupReadWriteExecute", func(t *testing.T) {
		file, err := os.Create(testFilePath)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFilePath)
		defer file.Close()

		// Set the file permissions
		err = file.Chmod(os.FileMode(UserReadWriteExecute | GroupReadWriteExecute))
		if err != nil {
			t.Fatalf("Failed to set file permissions: %v", err)
		}

		// Get the file permissions
		fileInfo, err := os.Stat(testFilePath)
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}
		// Check if the file permissions match the defined constants
		if fileInfo.Mode().Perm().String() != "-rwxrwx---" || fileInfo.Mode().Perm() != os.FileMode(UserReadWriteExecute|GroupReadWriteExecute) {
			t.Errorf("File permissions do not match. Expected: %s, Actual: %s", os.FileMode(UserReadWriteExecute|GroupReadWriteExecute).String(), fileInfo.Mode().Perm().String())
		}
	})
}
