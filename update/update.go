package updateutils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/charmbracelet/glamour"
	"github.com/minio/selfupdate"
	"github.com/projectdiscovery/gologger"
	errorutil "github.com/projectdiscovery/utils/errors"
)

const (
	Organization        = "projectdiscovery"
	UpdateCheckEndpoint = "https://api.pdtm.sh/api/v1/tools/%v"
)

var (
	// By default when tool is updated release notes of latest version are printed
	HideReleaseNotes      = false
	HideProgressBar       = false
	VersionCheckTimeout   = time.Duration(5) * time.Second
	DownloadUpdateTimeout = time.Duration(30) * time.Second
	// Note: DefaultHttpClient is only used in VersionCheck Callback
	DefaultHttpClient *http.Client
)

// GetUpdateToolCallback returns a callback function
// that updates given tool if given version is older than latest gh release and exits
func GetUpdateToolCallback(toolName, version string) func() {
	return func() {
		gh := NewghReleaseDownloader(toolName)
		latest, err := gh.GetLatestRelease()
		if err != nil {
			gologger.Fatal().Label("updater").Msgf("failed to fetch latest release of %v", toolName)
		}
		latestVersion, err := semver.NewVersion(latest.GetTagName())
		if err != nil {
			gologger.Fatal().Label("updater").Msgf("failed to parse semversion from tagname `%v` got %v", latest.GetTagName(), err)
		}
		currentVersion, err := semver.NewVersion(version)
		if err != nil {
			gologger.Fatal().Label("updater").Msgf("failed to parse semversion from current version %v got %v", version, err)
		}
		if !latestVersion.GreaterThan(currentVersion) {
			gologger.Info().Msgf("%v is already updated to latest version", toolName)
			os.Exit(0)
		}
		// check permissions before downloading release
		updateOpts := selfupdate.Options{}
		// TODO: selfupdate(https://github.com/minio/selfupdate) has support for checksum validation , code signing verification etc. implement them after discussion
		if err := updateOpts.CheckPermissions(); err != nil {
			gologger.Fatal().Label("updater").Msgf("update of %v %v -> %v failed , insufficient permission detected got: %v", toolName, currentVersion.String(), latestVersion.String(), err)
		}

		if err := gh.GetAssetIDFromRelease(latest); err != nil {
			gologger.Fatal().Label("updater").Msgf("failed to find release of %v for platform %v %v got : %v", toolName, runtime.GOOS, runtime.GOARCH, err)
		}
		bin, err := gh.GetExecutableFromAsset()
		if err != nil {
			gologger.Fatal().Label("updater").Msgf("executable %v not found in release asset `%v` got: %v", toolName, gh.AssetID, err)
		}

		if err = selfupdate.Apply(bytes.NewBuffer(bin), updateOpts); err != nil {
			log.Printf("Error] update of %v %v -> %v failed, rolling back update", toolName, currentVersion.String(), latestVersion.String())
			if err := selfupdate.RollbackError(err); err != nil {
				gologger.Fatal().Label("updater").Msgf("rollback of update of %v failed got %v,pls reinstall %v", toolName, err, toolName)
			}
			os.Exit(1)
		}

		gologger.Print().Msg("")
		gologger.Info().Msgf("%v sucessfully updated %v -> %v (latest)", toolName, currentVersion.String(), latestVersion.String())

		if !HideReleaseNotes {
			output := latest.GetBody()
			if rendered, err := glamour.Render(output, "dark"); err == nil {
				output = rendered
			} else {
				gologger.Error().Msg(err.Error())
			}
			gologger.Print().Msgf("%v\n\n", output)
		}
		os.Exit(0)
	}
}

// GetVersionCheckCallback retuns a callback function
// that returns latest version of tool
func GetVersionCheckCallback(toolName string) func() (string, error) {
	return func() (string, error) {
		updateURL := fmt.Sprintf(UpdateCheckEndpoint, toolName) + "?" + getMetaParams()
		if DefaultHttpClient == nil {
			// not needed but as a precaution to avoid nil panics
			DefaultHttpClient = http.DefaultClient
			DefaultHttpClient.Timeout = VersionCheckTimeout
		}
		resp, err := DefaultHttpClient.Get(updateURL)
		if err != nil {
			return "", errorutil.NewWithErr(err).Msgf("http Get %v failed", updateURL).WithTag("updater")
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		if resp.StatusCode != 200 {
			return "", errorutil.NewWithTag("updater", "version check failed expected status 200 but got %v for GET %v", resp.StatusCode, updateURL)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", errorutil.NewWithErr(err).Msgf("failed to get response body of GET %v", updateURL).WithTag("updater")
		}
		var toolDetails Tool
		if err := json.Unmarshal(body, &toolDetails); err != nil {
			return "", errorutil.NewWithErr(err).Msgf("failed to unmarshal %v", string(body)).WithTag("updater")
		}
		if toolDetails.Version == "" {
			return "", errorutil.NewWithErr(err).Msgf("something went wrong, expected version string but got empty string for GET `%v`", updateURL)
		}
		return toolDetails.Version, nil
	}
}

// getMetaParams returns encoded query parameters sent to update check endpoint
func getMetaParams() string {
	params := &url.Values{}
	params.Add("os", runtime.GOOS)
	params.Add("arch", runtime.GOARCH)
	params.Add("go_version", runtime.Version())
	return params.Encode()
}

func init() {
	DefaultHttpClient = &http.Client{
		Timeout: VersionCheckTimeout,
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
