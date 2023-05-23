package healthcheck

import (
	"os"
	"runtime"

	"github.com/projectdiscovery/fdmax"
	iputil "github.com/projectdiscovery/utils/ip"
	permissionutil "github.com/projectdiscovery/utils/permission"
	router "github.com/projectdiscovery/utils/routing"
)

type EnvironmentInfo struct {
	ExternalIPv4   string
	Admin          bool
	Arch           string
	Compiler       string
	GoVersion      string
	OSName         string
	ProgramVersion string
	OutboundIPv4   string
	OutboundIPv6   string
	Ulimit         Ulimit
	PathEnvVar     string
}

type Ulimit struct {
	Current uint64
	Max     uint64
}

func CollectEnvironmentInfo(appVersion string) (*EnvironmentInfo, error) {
	externalIPv4, _ := iputil.WhatsMyIP()
	outboundIPv4, outboundIPv6, _ := router.GetOutboundIPs()
	limit, _ := fdmax.Get()
	return &EnvironmentInfo{
		ExternalIPv4:   externalIPv4,
		Admin:          permissionutil.IsRoot,
		Arch:           runtime.GOARCH,
		Compiler:       runtime.Compiler,
		GoVersion:      runtime.Version(),
		OSName:         runtime.GOOS,
		ProgramVersion: appVersion,
		OutboundIPv4:   outboundIPv4.String(),
		OutboundIPv6:   outboundIPv6.String(),
		Ulimit: Ulimit{
			Current: limit.Current,
			Max:     limit.Max,
		},
		PathEnvVar: os.Getenv("PATH"),
	}, nil
}
