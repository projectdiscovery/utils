package folderutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	fileutil "github.com/projectdiscovery/utils/file"
	mapsutil "github.com/projectdiscovery/utils/maps"
)

var (
	// Separator evaluated at runtime
	Separator = string(os.PathSeparator)
	// Remove source directory after successful sync
	RemoveSourceDirAfterSync = true
	// writeablePathCache is a cache of writeable paths
	writeablePathCache = mapsutil.SyncLockMap[string, struct{}]{
		Map: make(map[string]struct{}),
	}
)

const (
	UnixPathSeparator    = "/"
	WindowsPathSeparator = "\\"
)

// GetFiles within a folder
func GetFiles(root string) ([]string, error) {
	var matches []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
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

// IsWritable checks if a path is writable by attempting to create a temporary file.
// It caches writable paths to avoid unnecessary file operations.
func IsWritable(path string) bool {
	if _, ok := writeablePathCache.Get(path); ok {
		return true
	}
	if !fileutil.FolderExists(path) {
		return false
	}
	tmpfile, err := os.CreateTemp(path, "test")
	if err != nil {
		return false
	}
	_ = tmpfile.Close()
	_ = os.Remove(tmpfile.Name())
	_ = writeablePathCache.Set(path, struct{}{})
	return true
}

// SyncDirectory sync all files and non-empty directories from source to destination folder
// optionally removes source directory and removes source
func SyncDirectory(source, destination string) error {
	// trim trailing slash to avoid slash related issues
	source = strings.TrimSuffix(source, Separator)
	destination = strings.TrimSuffix(destination, Separator)

	if !fileutil.FolderExists(source) {
		return errors.New("source directory doesn't exist")
	}

	if fileutil.FolderExists(destination) {
		sourceStat, err := os.Stat(source)
		if err != nil {
			return err
		}
		destinationStat, err := os.Stat(destination)
		if err != nil {
			return err
		}
		if os.SameFile(sourceStat, destinationStat) {
			return errors.New("source and destination cannot be the same")
		}
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destPath := filepath.Join(destination, entry.Name())

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

				err = SyncDirectory(sourcePath, destPath)
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

	if RemoveSourceDirAfterSync {
		err = os.RemoveAll(source)
		if err != nil {
			return err
		}
	}

	return nil
}

// DedupeLinesInFiles deduplicates lines in all files in a directory
// The function can be memory intensive for directories with large files.
func DedupeLinesInFiles(dir string) error {
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return fileutil.DedupeLines(path)
		}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "error processing directory %s", dir)
	}
	return nil
}
