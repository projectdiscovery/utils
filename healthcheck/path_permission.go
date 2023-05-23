package healthcheck

import (
	"errors"
	"path/filepath"

	fileutil "github.com/projectdiscovery/utils/file"
	folderutil "github.com/projectdiscovery/utils/folder"
)

var DefaultPathsToCheckPermission = []string{
	filepath.Join(folderutil.HomeDirOrDefault(""), ".config", fileutil.ExecutableName()),
}

type PathPermission struct {
	path       string
	isReadable bool
	isWritable bool
}

// CheckPathPermission checks the permissions of the given file or directory.
func CheckPathPermission(path string) (*PathPermission, error) {
	if !fileutil.FileExists(path) {
		return nil, errors.New("file or directory doesn't exist at " + path)
	}

	pathIsReadable, _ := fileutil.IsReadable(path)
	pathIsWritable, _ := fileutil.IsWriteable(path)

	return &PathPermission{
		path:       path,
		isReadable: pathIsReadable,
		isWritable: pathIsWritable,
	}, nil
}

// CheckPathsPermissionOrDefault checks the permissions of the given files or directories, or default files or directories if none are given.
func CheckPathsPermissionOrDefault(paths []string) ([]PathPermission, error) {
	if len(paths) == 0 {
		paths = DefaultPathsToCheckPermission
	}

	pathPermissions := []PathPermission{}
	for _, path := range paths {
		pathPermission, err := CheckPathPermission(path)
		if err != nil {
			return nil, err
		}
		pathPermissions = append(pathPermissions, *pathPermission)
	}
	return pathPermissions, nil
}
