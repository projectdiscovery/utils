package chromeshell

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestZip(path string, entries map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	w := zip.NewWriter(f)
	for name, content := range entries {
		e, err := w.Create(name)
		if err != nil {
			return err
		}
		if _, err := e.Write([]byte(content)); err != nil {
			return err
		}
	}
	return w.Close()
}

func TestHost(t *testing.T) {
	u := Host()
	if !strings.Contains(u, "chrome-for-testing-public/"+Version+"/") {
		t.Fatalf("unexpected host url: %s", u)
	}
	if !strings.HasSuffix(u, "/linux64/chrome-headless-shell-linux64.zip") {
		t.Fatalf("unexpected zip path: %s", u)
	}
	if HostChromeShell(0) != u {
		t.Fatalf("HostChromeShell mismatch")
	}
}

func TestFindBin(t *testing.T) {
	dir := t.TempDir()
	if got := findBin(dir); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}

	shell := filepath.Join(dir, "chrome-headless-shell")
	if err := os.WriteFile(shell, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := findBin(dir); got != shell {
		t.Fatalf("got %q want %q", got, shell)
	}
}

func TestUnzipRejectsPathEscape(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "evil.zip")
	if err := writeTestZip(zipPath, map[string]string{
		"../evil": "x",
	}); err != nil {
		t.Fatal(err)
	}
	if err := unzip(zipPath, filepath.Join(dir, "out")); err == nil {
		t.Fatal("expected error for path escape")
	}
}

func TestUnzipExtractsLocalEntries(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "ok.zip")
	if err := writeTestZip(zipPath, map[string]string{
		"chrome-headless-shell": "x",
	}); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	if err := unzip(zipPath, out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "chrome-headless-shell")); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureUnsupported(t *testing.T) {
	if Supported() {
		t.Skip("linux/amd64 supports download")
	}
	if _, err := Ensure(); err == nil {
		t.Fatal("expected unsupported error")
	}
}
