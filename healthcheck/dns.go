package healthcheck

import (
	"context"
	"net"
	"strings"
)

type DnsResolveInfo struct {
	Host        string
	Resolver    string
	Successful  bool
	IPAddresses []net.IPAddr
}

func DnsResolve(host string, resolver string) (*DnsResolveInfo, error) {
	ipAddresses, err := getIPAddresses(host, resolver)
	if err != nil {
		return nil, err
	}

	return &DnsResolveInfo{
		Host:        host,
		Resolver:    resolver,
		Successful:  true,
		IPAddresses: ipAddresses,
	}, nil
}

func getIPAddresses(name, dnsServer string) ([]net.IPAddr, error) {
	if !strings.Contains(dnsServer, ":") {
		dnsServer = dnsServer + ":53"
	}

	resolver := net.Resolver{
		PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, dnsServer)
		}}

	resolvedIPs, err := resolver.LookupIPAddr(context.Background(), name)
	if err != nil {
		return nil, err
	}

	return resolvedIPs, nil
}
