package chromeshell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestSafeJoin(t *testing.T) {
	base := t.TempDir()
	if _, err := safeJoin(base, "../evil"); err == nil {
		t.Fatal("expected error for path escape")
	}
	got, err := safeJoin(base, "chrome-headless-shell")
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(base, "chrome-headless-shell") {
		t.Fatalf("got %q", got)
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
