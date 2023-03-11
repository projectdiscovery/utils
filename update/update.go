package updateutils

import (
	"bytes"
	"log"
	"os"
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/minio/selfupdate"
)

// Organization/Owner Name
const Organization = "projectdiscovery"

// GetUpdateToolCallback returns a callback function
// that updates given tool if given version is older than latest gh release and exits
func GetUpdateToolCallback(toolName, version string) func() {
	return func() {
		log.SetFlags(0)

		gh := NewghReleaseDownloader(toolName)
		latest, err := gh.GetLatestRelease()
		if err != nil {
			log.Fatalf("updater: failed to fetch latest release of %v", toolName)
		}
		latestVersion, err := semver.NewVersion(latest.GetTagName())
		if err != nil {
			log.Fatalf("updater: failed to parse semversion from tagname `%v` got %v", latest.GetTagName(), err)
		}
		currentVersion, err := semver.NewVersion(version)
		if err != nil {
			log.Fatalf("updater: failed to parse semversion from current version %v got %v", version, err)
		}
		if !latestVersion.GreaterThan(currentVersion) {
			log.Printf("updater: %v is already updated to latest version", toolName)
		}
		// check permissions before downloading release
		updateOpts := selfupdate.Options{}
		// TODO: selfupdate(https://github.com/minio/selfupdate) has support for checksum validation , code signing verification etc. implement them after discussion
		if err := updateOpts.CheckPermissions(); err != nil {
			log.Fatalf("updater: [Error] update of %v %v -> %v failed , insufficient permission detected got: %v", toolName, currentVersion.String(), latestVersion.String(), err)
		}

		if err := gh.GetAssetIDFromRelease(latest); err != nil {
			log.Fatalf("updater: failed to find release of %v for platform %v %v got : %v", toolName, runtime.GOOS, runtime.GOARCH, err)
		}
		bin, err := gh.GetExecutableFromAsset()
		if err != nil {
			log.Fatalf("updater: executable %v not found in release asset `%v` got: %v", toolName, gh.AssetID, err)
		}

		if err = selfupdate.Apply(bytes.NewBuffer(bin), updateOpts); err != nil {
			log.Printf("updater: [Error] update of %v %v -> %v failed, rolling back update", toolName, currentVersion.String(), latestVersion.String())
			if err := selfupdate.RollbackError(err); err != nil {
				log.Fatalf("updater: rollback of update of %v failed got %v,pls reinstall %v", toolName, err, toolName)
			}
			os.Exit(1)
		}

		log.Printf("%v sucessfully updated %v -> %v (latest)", toolName, currentVersion.String(), latestVersion.String())
		os.Exit(0)
	}
}
