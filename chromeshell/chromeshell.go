package chromeshell

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	// Version is the pinned Chrome for Testing chrome-headless-shell build.
	// Smaller and faster than full Chromium for headless screenshots on Linux.
	Version = "152.0.7977.42"

	// revision is a synthetic cache key under the rod browser cache layout so
	// existing go-rod consumers share the same on-disk location.
	revision = 1520797742
)

var ensureMu sync.Mutex

// Supported reports whether chrome-headless-shell auto-download is available
// for the current platform.
func Supported() bool {
	return runtime.GOOS == "linux" && runtime.GOARCH == "amd64"
}

// Host returns the Chrome for Testing chrome-headless-shell zip URL for
// linux/amd64. Other platforms should not use this host for downloads.
func Host() string {
	return fmt.Sprintf(
		"https://storage.googleapis.com/chrome-for-testing-public/%s/linux64/chrome-headless-shell-linux64.zip",
		Version,
	)
}

// HostChromeShell is a go-rod launcher.Host-compatible adapter that ignores the
// revision argument and always returns the pinned chrome-headless-shell URL.
func HostChromeShell(_ int) string {
	return Host()
}

// Dir returns the on-disk cache directory for the pinned chrome-headless-shell.
func Dir() string {
	return filepath.Join(defaultBrowserDir(), fmt.Sprintf("chromium-%d", revision))
}

// BinPath returns the preferred executable path inside Dir, if present.
func BinPath() string {
	return findBin(Dir())
}

// Ensure downloads chrome-headless-shell once into the shared browser cache and
// returns the executable path. It is a no-op download when the binary already
// exists. Callers should only invoke this when Supported() is true.
func Ensure() (string, error) {
	if !Supported() {
		return "", fmt.Errorf("chrome-headless-shell auto-download is only supported on linux/amd64")
	}

	ensureMu.Lock()
	defer ensureMu.Unlock()

	if p := findBin(Dir()); p != "" {
		return p, nil
	}

	if err := downloadAndExtract(Host(), Dir()); err != nil {
		return "", err
	}

	p := findBin(Dir())
	if p == "" {
		return "", fmt.Errorf("chrome-headless-shell binary missing after download in %s", Dir())
	}

	// go-rod's linux BinPath expects "chrome"; keep a symlink for Validate().
	chrome := filepath.Join(Dir(), "chrome")
	if p != chrome {
		_ = os.Remove(chrome)
		_ = os.Symlink(filepath.Base(p), chrome)
	}

	return p, nil
}

func findBin(dir string) string {
	for _, name := range []string{"chrome-headless-shell", "headless_shell", "chrome"} {
		p := filepath.Join(dir, name)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	return ""
}

func defaultBrowserDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.TempDir()
	}
	switch runtime.GOOS {
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "rod", "browser")
		}
	}
	return filepath.Join(home, ".cache", "rod", "browser")
}

func downloadAndExtract(url, destDir string) error {
	tmpParent := filepath.Dir(destDir)
	if err := os.MkdirAll(tmpParent, 0o755); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp(tmpParent, "chrome-headless-shell-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	zipPath := filepath.Join(tmpDir, "chrome-headless-shell.zip")
	if err := downloadFile(url, zipPath); err != nil {
		return err
	}

	extractDir := filepath.Join(tmpDir, "extract")
	if err := unzip(zipPath, extractDir); err != nil {
		return err
	}

	// Chrome for Testing zips nest files under one top-level directory.
	contentDir, err := stripFirstDir(extractDir)
	if err != nil {
		return err
	}

	_ = os.RemoveAll(destDir)
	if err := os.Rename(contentDir, destDir); err != nil {
		return err
	}
	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url) //nolint:noctx // one-shot browser binary fetch
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: unexpected status %s", url, resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, resp.Body)
	return err
}

func unzip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	for _, f := range r.File {
		// entry names must stay inside dest: reject "..", absolute paths and
		// other non-local names before joining
		if !filepath.IsLocal(f.Name) {
			return fmt.Errorf("illegal path in zip: %s", f.Name)
		}
		target := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := writeZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func writeZipFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	mode := f.Mode()
	if mode&0o111 != 0 {
		mode = 0o755
	} else {
		mode = 0o644
	}

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, rc)
	return err
}

func stripFirstDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(dir, e.Name()))
		}
	}
	if len(dirs) == 1 && len(entries) == 1 {
		return dirs[0], nil
	}
	return dir, nil
}
