package updateutils

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/logrusorgru/aurora"
)

type AssetFormat uint

const (
	Zip AssetFormat = iota
	Tar
)

// Tool
type Tool struct {
	Name    string            `json:"name"`
	Repo    string            `json:"repo"`
	Version string            `json:"version"`
	Assets  map[string]string `json:"assets"`
}

// GetVersionDescription returns tags like (latest) or (outdated) or (dev)
func GetVersionDescription(current string, latest string) string {
	currentVer, _ := semver.NewVersion(current)
	latestVer, _ := semver.NewVersion(latest)
	if strings.Contains(current, "dev") {
		return fmt.Sprintf("(%v)", aurora.BrightBlue("dev"))
	}
	if currentVer == nil || latestVer == nil {
		// fallback to naive comparison
		if current == latest {
			return fmt.Sprintf("(%v)", aurora.BrightGreen("latest"))
		} else {
			return fmt.Sprintf("(%v)", aurora.BrightRed("outdated"))
		}
	}
	if latestVer.GreaterThan(currentVer) {
		return fmt.Sprintf("(%v)", aurora.BrightRed("outdated"))
	} else {
		return fmt.Sprintf("(%v)", aurora.BrightGreen("latest"))
	}
}
