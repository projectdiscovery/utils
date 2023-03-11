package updateutils

import (
	"context"
	"log"
	"os"

	"github.com/creativeprojects/go-selfupdate"
)

// Organization/Owner Name
const Organization = "projectdiscovery"

// GetUpdateToolCallback returns a callback function
// that updates given tool if given version is older than latest gh release and exits
func GetUpdateToolCallback(toolName, version string) func() {
	return func() {
		latest, ok, err := selfupdate.DetectLatest(context.TODO(), selfupdate.NewRepositorySlug(Organization, toolName))
		if !ok {
			log.Fatalf("failed to fetch latest version of %v got %v error", toolName, err)
		}
		if latest.LessOrEqual(version) {
			log.Printf("current version of %v is latest", toolName)
			os.Exit(0)
		}
		toolPath, err := os.Executable()
		if err != nil {
			log.Fatalf("could not find path of %v. exiting", toolName)
		}
		if err := selfupdate.UpdateTo(context.TODO(), latest.AssetURL, latest.AssetName, toolPath); err != nil {
			log.Fatalf("failed to update %v to latest version got %v error", toolName, err)
		}
		log.Printf("Successfully updated %v to version %v", toolName, latest.Version())
		os.Exit(0)
	}
}
