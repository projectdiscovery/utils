package dnsutil

import (
	stringsutil "github.com/projectdiscovery/utils/strings"
	"github.com/weppos/publicsuffix-go/publicsuffix"
)

// DecomposeDomain decomposes a given domain into subdomain and domain.
// It returns the subdomain and the domain.
func DecomposeDomain(name string) (string, string, error) {
	name = stringsutil.TrimPrefixAny(name, "http://", "https://")
	dn, err := publicsuffix.ParseFromListWithOptions(publicsuffix.DefaultList, name, publicsuffix.DefaultFindOptions)
	if err != nil {
		return "", "", err
	}

	return dn.TRD, dn.SLD + "." + dn.TLD, nil
}
