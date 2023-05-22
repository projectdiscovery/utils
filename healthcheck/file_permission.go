package healthcheck

import (
	"errors"
	"path/filepath"

	fileutil "github.com/projectdiscovery/utils/file"
	folderutil "github.com/projectdiscovery/utils/folder"
)

var DefaultFilesToCheckPermissions = []string{filepath.Join(folderutil.HomeDirOrDefault(""), ".config")}

type FilePermissions struct {
	filename   string
	isReadable bool
	isWritable bool
}

// CheckFilePermissions checks the permissions of the given file.
func CheckFilePermissions(filename string) (*FilePermissions, error) {
	if !fileutil.FileExists(filename) {
		return nil, errors.New("file doesn't exist")
	}

	fileIsReadable, _ := fileutil.IsReadable(filename)
	fileIsWritable, _ := fileutil.IsWriteable(filename)

	return &FilePermissions{
		filename:   filename,
		isReadable: fileIsReadable,
		isWritable: fileIsWritable,
	}, nil
}

// CheckFilesPermissionsOrDefault checks the permissions of the given files or default files if none are given.
func CheckFilesPermissionsOrDefault(filenames []string) ([]FilePermissions, error) {
	if len(filenames) == 0 {
		filenames = DefaultFilesToCheckPermissions
	}

	filePermissions := []FilePermissions{}
	for _, filename := range filenames {
		filePermission, err := CheckFilePermissions(filename)
		if err != nil {
			return nil, err
		}
		filePermissions = append(filePermissions, *filePermission)
	}
	return filePermissions, nil
}
