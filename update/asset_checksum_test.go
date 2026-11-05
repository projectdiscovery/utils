package updateutils

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-github/v92/github"
	"github.com/stretchr/testify/require"
)

func mkzip(t *testing.T, name string, content []byte) []byte {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write(content)
	_ = zw.Close()
	return b.Bytes()
}

// fakeGH serves a GitHub-like API: /repos/o/r/releases/latest and
// /repos/o/r/releases/assets/<id> (redirect to /dl/<id>).
func fakeGH(t *testing.T, tag string, assets map[string][]byte) *httptest.Server {
	t.Helper()
	names := []string{}
	for n := range assets {
		names = append(names, n)
	}
	mux := http.NewServeMux()
	var srv *httptest.Server
	mux.HandleFunc("/repos/o/tool/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		var parts []string
		for i, n := range names {
			parts = append(parts, fmt.Sprintf(`{"id":%d,"name":%q}`, i+1, n))
		}
		_, _ = fmt.Fprintf(w, `{"tag_name":%q,"assets":[%s]}`, tag, strings.Join(parts, ","))
	})
	mux.HandleFunc("/repos/o/tool/releases/assets/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/repos/o/tool/releases/assets/")
		http.Redirect(w, r, srv.URL+"/dl/"+id, http.StatusFound)
	})
	mux.HandleFunc("/dl/", func(w http.ResponseWriter, r *http.Request) {
		var i int
		_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/dl/"), "%d", &i)
		_, _ = w.Write(assets[names[i-1]])
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func newDL(t *testing.T, srv *httptest.Server) *GHReleaseDownloader {
	hc := &http.Client{}
	base := srv.URL + "/"
	c, err := github.NewClient(github.WithHTTPClient(hc), github.WithURLs(&base, &base))
	if err != nil {
		t.Fatal(err)
	}
	d := &GHReleaseDownloader{client: c, repoName: "tool", assetName: "tool", httpClient: hc, organization: "o"}
	if err := d.getLatestRelease(); err != nil {
		t.Fatal(err)
	}
	return d
}

func assetName(tag string) string {
	osn := runtime.GOOS
	if osn == "darwin" {
		osn = "macOS"
	}
	return fmt.Sprintf("tool_%s_%s_%s.zip", strings.TrimPrefix(tag, "v"), osn, runtime.GOARCH)
}

func checksumLine(t *testing.T, name string, data []byte) []byte {
	t.Helper()
	sum := sha256.Sum256(data)
	return []byte(hex.EncodeToString(sum[:]) + "  " + name + "\n")
}

func TestGetExecutableFromAssetVerifiesChecksum(t *testing.T) {
	HideProgressBar = true
	tag := "v9.9.9"
	z := mkzip(t, "tool", []byte("BINARY"))

	t.Run("no checksums file", func(t *testing.T) {
		d := newDL(t, fakeGH(t, tag, map[string][]byte{assetName(tag): z}))
		_, err := d.GetExecutableFromAsset()
		require.Error(t, err)
	})
	t.Run("no entry for the asset", func(t *testing.T) {
		d := newDL(t, fakeGH(t, tag, map[string][]byte{
			assetName(tag):             z,
			"tool_9.9.9_checksums.txt": []byte("deadbeef  some_other_asset.zip\n"),
		}))
		_, err := d.GetExecutableFromAsset()
		require.Error(t, err)
	})
	t.Run("checksum mismatch", func(t *testing.T) {
		d := newDL(t, fakeGH(t, tag, map[string][]byte{
			assetName(tag):             z,
			"tool_9.9.9_checksums.txt": checksumLine(t, assetName(tag), []byte("other")),
		}))
		_, err := d.GetExecutableFromAsset()
		require.Error(t, err)
	})
	for _, name := range []string{"tool_9.9.9_checksums.txt", "tool-9.9.9-checksums.txt", "tool-checksums.txt"} {
		t.Run("valid "+name, func(t *testing.T) {
			d := newDL(t, fakeGH(t, tag, map[string][]byte{
				assetName(tag): z,
				name:           checksumLine(t, assetName(tag), z),
			}))
			bin, err := d.GetExecutableFromAsset()
			require.NoError(t, err)
			require.Equal(t, "BINARY", string(bin))
		})
	}
	t.Run("skip validation", func(t *testing.T) {
		SkipCheckSumValidation = true
		defer func() { SkipCheckSumValidation = false }()
		d := newDL(t, fakeGH(t, tag, map[string][]byte{assetName(tag): z}))
		bin, err := d.GetExecutableFromAsset()
		require.NoError(t, err)
		require.Equal(t, "BINARY", string(bin))
	})
}

func TestGetExecutableFromAssetWithoutBinary(t *testing.T) {
	HideProgressBar = true
	tag := "v9.9.9"
	z := mkzip(t, "README.md", []byte("no binary here"))
	d := newDL(t, fakeGH(t, tag, map[string][]byte{
		assetName(tag):             z,
		"tool_9.9.9_checksums.txt": checksumLine(t, assetName(tag), z),
	}))
	bin, err := d.GetExecutableFromAsset()
	require.Error(t, err)
	require.Empty(t, bin)
}
