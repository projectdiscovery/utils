package wildcard

import (
	"net"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// RegistrableRoot returns the registrable root for the provided host.
// The input is normalized first by lowercasing it, trimming a leading *., and
// removing any trailing dot. IP literals and invalid hosts are rejected.
func RegistrableRoot(host string) (string, bool) {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimPrefix(host, "*.")
	host = strings.TrimSuffix(host, ".")
	if host == "" {
		return "", false
	}

	if ip := net.ParseIP(host); ip != nil {
		return "", false
	}

	root, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil || root == "" {
		return "", false
	}

	return root, true
}
