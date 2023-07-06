package folderutil

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	fileutil "github.com/projectdiscovery/utils/file"
)

// Separator evaluated at runtime
var Separator = string(os.PathSeparator)

const (
	UnixPathSeparator    = "/"
	WindowsPathSeparator = "\\"
)

// GetFiles within a folder
func GetFiles(root string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		matches = append(matches, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// PathInfo about a folder
type PathInfo struct {
	IsAbsolute         bool
	RootPath           string
	Parts              []string
	PartsWithSeparator []string
}

// NewPathInfo returns info about a given path
func NewPathInfo(path string) (PathInfo, error) {
	var pathInfo PathInfo
	path = filepath.Clean(path)
	pathInfo.RootPath = filepath.VolumeName(path)
	if filepath.IsAbs(path) {
		if IsUnixOS() {
			if pathInfo.RootPath == "" {
				pathInfo.IsAbsolute = true
				pathInfo.RootPath = UnixPathSeparator
			}
		} else if IsWindowsOS() {
			pathInfo.IsAbsolute = true
			pathInfo.RootPath = pathInfo.RootPath + WindowsPathSeparator
		}
	}

	pathInfo.Parts = agnosticSplit(path)

	for i, pathItem := range pathInfo.Parts {
		if i == 0 && pathInfo.IsAbsolute {
			if IsUnixOS() {
				pathInfo.PartsWithSeparator = append(pathInfo.PartsWithSeparator, pathInfo.RootPath)
			}
		} else if len(pathInfo.PartsWithSeparator) > 0 && pathInfo.PartsWithSeparator[len(pathInfo.PartsWithSeparator)-1] != Separator {
			pathInfo.PartsWithSeparator = append(pathInfo.PartsWithSeparator, Separator)
		}
		pathInfo.PartsWithSeparator = append(pathInfo.PartsWithSeparator, pathItem)
	}
	return pathInfo, nil
}

// Returns all possible combination of the various levels of the path parts
func (pathInfo PathInfo) Paths() ([]string, error) {
	var combos []string
	for i := 0; i <= len(pathInfo.Parts); i++ {
		var computedPath string
		if pathInfo.IsAbsolute && pathInfo.RootPath != "" {
			// on windows we need to skip the volume, already computed in rootpath
			if IsUnixOS() {
				computedPath = pathInfo.RootPath + filepath.Join(pathInfo.Parts[:i]...)
			} else if IsWindowsOS() && i > 0 {
				skipItems := 0
				if len(pathInfo.Parts) > 0 {
					skipItems = 1
				}
				computedPath = pathInfo.RootPath + filepath.Join(pathInfo.Parts[skipItems:i]...)
			}
		} else {
			computedPath = filepath.Join(pathInfo.Parts[:i]...)
		}
		combos = append(combos, filepath.Clean(computedPath))
	}

	return combos, nil
}

// MeshWith combine all values from Path with another provided path
func (pathInfo PathInfo) MeshWith(anotherPath string) ([]string, error) {
	allPaths, err := pathInfo.Paths()
	if err != nil {
		return nil, err
	}
	var combos []string
	for _, basePath := range allPaths {
		combinedPath := filepath.Join(basePath, anotherPath)
		combos = append(combos, filepath.Clean(combinedPath))
	}

	return combos, nil
}

func IsUnixOS() bool {
	switch runtime.GOOS {
	case "android", "darwin", "freebsd", "ios", "linux", "netbsd", "openbsd", "solaris":
		return true
	default:
		return false
	}
}

func IsWindowsOS() bool {
	return runtime.GOOS == "windows"
}

func agnosticSplit(path string) (parts []string) {
	// split with each known separators
	for _, part := range strings.Split(path, UnixPathSeparator) {
		for _, subPart := range strings.Split(part, WindowsPathSeparator) {
			if part != "" {
				parts = append(parts, subPart)
			}
		}
	}
	return
}

// HomeDirectory returns the home directory or defaultDirectory in case of error
func HomeDirOrDefault(defaultDirectory string) string {
	usr, err := user.Current()
	if err != nil {
		return defaultDirectory
	}
	return usr.HomeDir
}

// UserConfigDirOrDefault returns the user config directory or defaultConfigDir in case of error
func UserConfigDirOrDefault(defaultConfigDir string) string {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return defaultConfigDir
	}
	return userConfigDir
}

// AppConfigDirOrDefault returns the app config directory
func AppConfigDirOrDefault(defaultAppConfigDir string, toolName string) string {
	userConfigDir := UserConfigDirOrDefault("")
	if userConfigDir == "" {
		return filepath.Join(defaultAppConfigDir, toolName)
	}
	return filepath.Join(userConfigDir, toolName)
}

// MigrateDir moves all files and non-empty directories from sourceDir to destinationDir and removes sourceDir
func MigrateDir(sourceDir string, destinationDir string, removeSourceDir bool) error {
	// trim trailing slash to avoid slash related issues
	sourceDir = strings.TrimSuffix(sourceDir, Separator)
	destinationDir = strings.TrimSuffix(destinationDir, Separator)

	if !fileutil.FolderExists(sourceDir) {
		return errors.New("source directory doesn't exist")
	}

	if fileutil.FolderExists(destinationDir) {
		sourceStat, err := os.Stat(sourceDir)
		if err != nil {
			return err
		}
		destinationStat, err := os.Stat(destinationDir)
		if err != nil {
			return err
		}
		if os.SameFile(sourceStat, destinationStat) {
			return errors.New("sourceDir and destinationDir cannot be the same")
		}
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		destPath := filepath.Join(destinationDir, entry.Name())

		if entry.IsDir() {
			subentries, err := os.ReadDir(sourcePath)
			if err != nil {
				return err
			}
			if len(subentries) > 0 {
				err = os.MkdirAll(destPath, os.ModePerm)
				if err != nil {
					return err
				}

				err = MigrateDir(sourcePath, destPath, removeSourceDir)
				if err != nil {
					return err
				}
			}
		} else {
			err = os.Rename(sourcePath, destPath)
			if err != nil {
				return err
			}
		}
	}

	if removeSourceDir {
		err = os.RemoveAll(sourceDir)
		if err != nil {
			return err
		}
	}

	return nil
}
