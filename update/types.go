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
	Unknown
)

// FileExtension of this asset format
func (a AssetFormat) FileExtension() string {
	if a == Zip {
		return ".zip"
	} else if a == Tar {
		return ".tar.gz"
	}
	return ""
}

func IdentifyAssetFormat(assetName string) AssetFormat {
	switch {
	case strings.HasSuffix(assetName, Zip.FileExtension()):
		return Zip
	case strings.HasSuffix(assetName, Tar.FileExtension()):
		return Tar
	default:
		return Unknown
	}
}

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
	if strings.HasSuffix(current, "-dev") {
		return fmt.Sprintf("(%v)", aurora.Blue("development"))
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
