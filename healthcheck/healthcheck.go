package healthcheck

import (
	"path/filepath"

	fileutil "github.com/projectdiscovery/utils/file"
	folderutil "github.com/projectdiscovery/utils/folder"
)

var (
	DefaultPathsToCheckPermission   = []string{filepath.Join(folderutil.HomeDirOrDefault(""), ".config", fileutil.ExecutableName())}
	DefaultHostsToCheckConnectivity = []string{"scanme.sh"}
	DefaultResolver                 = "1.1.1.1:53"
)

type HealthCheck struct {
	paths    []string
	hosts    []string
	resolver string
}

type HealthCheckInfo struct {
	EnvironmentInfo EnvironmentInfo
	PathPermissions []PathPermission
	DnsResolveInfos []DnsResolveInfo
}

func New(paths, hosts []string, resolver string) *HealthCheck {
	if len(paths) == 0 {
		paths = DefaultPathsToCheckPermission
	}
	if len(hosts) == 0 {
		hosts = DefaultHostsToCheckConnectivity
	}
	if resolver == "" {
		resolver = DefaultResolver
	}

	return &HealthCheck{paths: paths, hosts: hosts, resolver: resolver}
}

func (h *HealthCheck) Run(programVersion string) (*HealthCheckInfo, error) {
	environmentInfo, err := CollectEnvironmentInfo(programVersion)
	if err != nil {
		return nil, err
	}

	pathPermissions := []PathPermission{}
	for _, path := range h.paths {
		pathPermission, err := CheckPathPermission(path)
		if err != nil {
			return nil, err
		}
		pathPermissions = append(pathPermissions, *pathPermission)
	}

	dnsResolveInfos := []DnsResolveInfo{}
	for _, host := range h.hosts {
		dnsResolveInfo, err := DnsResolve(host, h.resolver)
		if err != nil {
			return nil, err
		}
		dnsResolveInfos = append(dnsResolveInfos, *dnsResolveInfo)
	}

	return &HealthCheckInfo{
		EnvironmentInfo: *environmentInfo,
		PathPermissions: pathPermissions,
		DnsResolveInfos: dnsResolveInfos,
	}, nil
}
