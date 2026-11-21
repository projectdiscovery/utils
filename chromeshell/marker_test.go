package chromeshell

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// archiveServer serves a chrome-headless-shell archive and counts requests.
func archiveServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	zipPath := filepath.Join(t.TempDir(), "shell.zip")
	if err := writeTestZip(zipPath, map[string]string{
		"chrome-headless-shell-linux64/chrome-headless-shell": "shell",
	}); err != nil {
		t.Fatal(err)
	}
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.ServeFile(w, r, zipPath)
	}))
	t.Cleanup(server.Close)
	return server, &hits
}

// closedURL returns a URL on a local port that refuses connections.
func closedURL(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return "http://" + addr + "/shell.zip"
}

func TestFailedDownloadWritesMarker(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "browser", "chromium-1")
	marker := dir + ".download-failed"

	if _, err := ensure(context.Background(), Options{}, closedURL(t), dir, marker); err == nil {
		t.Fatal("expected the download to fail")
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("marker not written: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("marker is empty")
	}
}

func TestMarkerSkipsDownloadWithin24h(t *testing.T) {
	server, hits := archiveServer(t)
	dir := filepath.Join(t.TempDir(), "browser", "chromium-1")
	marker := dir + ".download-failed"
	writeFailureMarker(marker, os.ErrDeadlineExceeded)

	start := time.Now()
	_, err := ensure(context.Background(), Options{}, server.URL, dir, marker)
	if err == nil || !strings.Contains(err.Error(), "skipped") {
		t.Fatalf("expected a skipped download, got %v", err)
	}
	if hits.Load() != 0 {
		t.Fatal("the download must not be tried within 24h of a failure")
	}
	if time.Since(start) > time.Second {
		t.Fatal("a skipped download must return at once")
	}
}

func TestMarkerRetriesAfter24h(t *testing.T) {
	server, hits := archiveServer(t)
	dir := filepath.Join(t.TempDir(), "browser", "chromium-1")
	marker := dir + ".download-failed"
	writeFailureMarker(marker, os.ErrDeadlineExceeded)
	old := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(marker, old, old); err != nil {
		t.Fatal(err)
	}

	p, err := ensure(context.Background(), Options{}, server.URL, dir, marker)
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 || p != filepath.Join(dir, "chrome-headless-shell") {
		t.Fatalf("expected one download and an installed binary, got hits=%d path=%q", hits.Load(), p)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("a successful download must remove the marker")
	}
}

// A canceled caller says nothing about the network, so it leaves no marker.
func TestCanceledDownloadWritesNoMarker(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "browser", "chromium-1")
	marker := dir + ".download-failed"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)
	if _, err := ensure(ctx, Options{}, server.URL, dir, marker); err == nil {
		t.Fatal("expected the canceled download to fail")
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("a canceled download must not write the marker")
	}
}

// The download goes through Options.Proxy. The archive host does not resolve,
// so the download works only when the proxy serves it.
func TestDownloadUsesProxy(t *testing.T) {
	archive, hits := archiveServer(t)
	var proxied atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Host != "chrome.invalid" {
			http.Error(w, "unexpected host "+r.URL.Host, http.StatusBadGateway)
			return
		}
		proxied.Add(1)
		resp, err := http.Get(archive.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}))
	t.Cleanup(proxy.Close)

	dir := filepath.Join(t.TempDir(), "browser", "chromium-1")
	p, err := ensure(context.Background(), Options{Proxy: proxy.URL}, "http://chrome.invalid/shell.zip", dir, dir+".download-failed")
	if err != nil {
		t.Fatal(err)
	}
	if proxied.Load() != 1 || hits.Load() != 1 || p == "" {
		t.Fatalf("expected the download through the proxy, got proxied=%d hits=%d path=%q", proxied.Load(), hits.Load(), p)
	}
}

func TestDownloadClientRejectsBadProxy(t *testing.T) {
	if _, err := downloadClient(Options{Proxy: "http://[::1"}); err == nil {
		t.Fatal("expected an invalid proxy error")
	}
}
