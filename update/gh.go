package updateutils

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/google/go-github/v30/github"
	errorutil "github.com/projectdiscovery/utils/errors"
	"golang.org/x/oauth2"
)

var (
	extIfFound      = ".exe"
	ErrNoAssetFound = errorutil.NewWithFmt("update: could not find release asset for your platform (%s/%s)")
)

// GHReleaseDownloader fetches and reads release of a gh repo
type GHReleaseDownloader struct {
	ToolName string // we assume toolname and ToolName are always same
	Format   AssetFormat
	AssetID  int

	client *github.Client
}

// NewghReleaseDownloader instance
func NewghReleaseDownloader(toolName string) *GHReleaseDownloader {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		DefaultHttpClient = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	}
	if DefaultHttpClient != nil {
		DefaultHttpClient.Timeout = DownloadUpdateTimeout
	}
	return &GHReleaseDownloader{client: github.NewClient(DefaultHttpClient), ToolName: toolName}
}

// getLatestRelease returns latest release of error
func (d *GHReleaseDownloader) GetLatestRelease() (*github.RepositoryRelease, error) {
	release, resp, err := d.client.Repositories.GetLatestRelease(context.Background(), Organization, d.ToolName)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, fmt.Errorf("repo %v/%v not found got %v", Organization, d.ToolName, err)
		}
		return nil, err
	}
	return release, nil
}

// getAssetIDFromRelease finds AssetID from release or returns a descriptive error
func (d *GHReleaseDownloader) GetAssetIDFromRelease(latest *github.RepositoryRelease) error {
	builder := &strings.Builder{}
	builder.WriteString(d.ToolName)
	builder.WriteString("_")
	builder.WriteString(strings.TrimPrefix(*latest.TagName, "v"))
	builder.WriteString("_")
	if strings.EqualFold(runtime.GOOS, "darwin") {
		builder.WriteString("macOS")
	} else {
		builder.WriteString(runtime.GOOS)
	}
	builder.WriteString("_")
	builder.WriteString(runtime.GOARCH)

loop:
	for _, v := range latest.Assets {
		asset := *v.Name
		switch {
		case strings.Contains(asset, ".zip"):
			if strings.EqualFold(asset, builder.String()+".zip") {
				d.AssetID = int(*v.ID)
				d.Format = Zip
				break loop
			}
		case strings.Contains(asset, ".tar.gz"):
			if strings.EqualFold(asset, builder.String()+".tar.gz") {
				d.AssetID = int(*v.ID)
				d.Format = Tar
				break loop
			}
		}
	}
	builder.Reset()

	// handle if id is zero (no asset found)
	if d.AssetID == 0 {
		return ErrNoAssetFound.Msgf(runtime.GOOS, runtime.GOARCH)
	}
	return nil
}

// DownloadAssetFromID downloads and returns a buffer or a descriptive error
func (d *GHReleaseDownloader) DownloadAssetFromID() (*bytes.Buffer, error) {
	_, rdurl, err := d.client.Repositories.DownloadReleaseAsset(context.Background(), Organization, d.ToolName, int64(d.AssetID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(rdurl)
	if err != nil {
		return nil, errorutil.NewWithErr(err).Msgf("failed to download release asset")
	}
	if resp.StatusCode != 200 {
		return nil, errorutil.New("something went wrong got %v while downloading asset, expected status 200", resp.StatusCode)
	}
	if resp.Body == nil {
		return nil, errorutil.New("something went wrong got response without body")
	}
	defer resp.Body.Close()

	if !HideProgressBar {
		bar := pb.New64(resp.ContentLength).SetMaxWidth(100)
		bar.Start()
		resp.Body = bar.NewProxyReader(resp.Body)
		defer bar.Finish()
	}

	bin, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errorutil.NewWithErr(err).Msgf("failed to read response body")
	}
	return bytes.NewBuffer(bin), nil
}

// GetExecutableFromAsset downloads and only returns tool Binary
func (d *GHReleaseDownloader) GetExecutableFromAsset() ([]byte, error) {
	buff, err := d.DownloadAssetFromID()
	if err != nil {
		return nil, err
	}
	if d.Format == Zip {
		zipReader, err := zip.NewReader(bytes.NewReader(buff.Bytes()), int64(buff.Len()))
		if err != nil {
			return nil, err
		}
		for _, f := range zipReader.File {
			if !strings.EqualFold(strings.TrimSuffix(f.Name, extIfFound), d.ToolName) {
				continue
			}
			fileInArchive, err := f.Open()
			if err != nil {
				return nil, err
			}
			bin, err := io.ReadAll(fileInArchive)
			if err != nil {
				return nil, err
			}
			_ = fileInArchive.Close()
			return bin, nil
		}
	} else if d.Format == Tar {
		gzipReader, err := gzip.NewReader(buff)
		if err != nil {
			return nil, err
		}
		tarReader := tar.NewReader(gzipReader)
		// iterate through the files in the archive
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if !strings.EqualFold(strings.TrimSuffix(header.FileInfo().Name(), extIfFound), d.ToolName) {
				continue
			}
			// if the file is not a directory, extract it
			if !header.FileInfo().IsDir() {
				bin, err := io.ReadAll(tarReader)
				if err != nil {
					return nil, err
				}
				return bin, nil
			}
		}
	}
	return nil, fmt.Errorf("executable not found in archive")
}
