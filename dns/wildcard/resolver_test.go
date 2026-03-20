package wildcard

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	sliceutil "github.com/projectdiscovery/utils/slice"
	"github.com/stretchr/testify/require"
)

func TestGenerateWildcardPermutations(t *testing.T) {
	var tests = []struct {
		subdomain string
		domain    string
		expected  []string
	}{
		{"test", "example.com", []string{"*.example.com"}},
		{"abc.test", "example.com", []string{"*.example.com", "*.test.example.com"}},
		{"xyz.abc.test", "example.com", []string{"*.example.com", "*.test.example.com", "*.abc.test.example.com"}},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s.%s", test.subdomain, test.domain), func(t *testing.T) {
			result := generateWildcardPermutations(test.subdomain, test.domain)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestResolverLookupHostSkipsUnknownDomain(t *testing.T) {
	resolver := NewResolver([]string{"example.com"}, func(host string) ([]string, error) {
		t.Fatalf("unexpected lookup for %s", host)
		return nil, nil
	})

	isWildcard, wildcards := resolver.LookupHost("www.other.com", []string{"1.1.1.1"})
	require.False(t, isWildcard)
	require.Nil(t, wildcards)
}

func TestResolverLookupHostReturnsCachedWildcardIPs(t *testing.T) {
	resolver := NewResolver([]string{"example.com"}, func(host string) ([]string, error) {
		return []string{"1.1.1.1", "2.2.2.2"}, nil
	})

	isWildcard, wildcards := resolver.LookupHost("www.example.com", []string{"1.1.1.1"})
	require.True(t, isWildcard)
	require.Equal(t, map[string]struct{}{"1.1.1.1": {}, "2.2.2.2": {}}, wildcards)
	require.Equal(t, wildcards, resolver.GetAllWildcardIPs())
}

func TestResolverLookupHostReturnsObservedWildcardIPsForNonMatch(t *testing.T) {
	resolver := NewResolver([]string{"example.com"}, func(host string) ([]string, error) {
		if host == "www.example.com" {
			return []string{"3.3.3.3"}, nil
		}
		return []string{"1.1.1.1", "2.2.2.2"}, nil
	})

	isWildcard, wildcards := resolver.LookupHost("www.example.com", []string{"3.3.3.3"})
	require.False(t, isWildcard)
	require.Equal(t, map[string]struct{}{"1.1.1.1": {}, "2.2.2.2": {}}, wildcards)
}

func TestResolverLookupHostReprobesCachedWildcard(t *testing.T) {
	count := 0
	resolver := NewResolver([]string{"example.com"}, func(host string) ([]string, error) {
		count++
		if count == 1 {
			return []string{"1.1.1.1"}, nil
		}
		return []string{"1.1.1.1", "2.2.2.2"}, nil
	})

	isWildcard, _ := resolver.LookupHost("www.example.com", []string{"1.1.1.1"})
	require.True(t, isWildcard)

	isWildcard, wildcards := resolver.LookupHost("api.example.com", []string{"2.2.2.2"})
	require.True(t, isWildcard)
	require.Contains(t, wildcards, "2.2.2.2")
}

func TestResolverLookupHostRevalidatesCurrentHost(t *testing.T) {
	hostLookups := 0
	resolver := NewResolver([]string{"example.com"}, func(host string) ([]string, error) {
		if host == "target.example.com" {
			hostLookups++
			if hostLookups < 3 {
				return []string{"3.3.3.3"}, nil
			}
			return []string{"2.2.2.2"}, nil
		}
		return []string{"2.2.2.2"}, nil
	})

	isWildcard, wildcards := resolver.LookupHost("target.example.com", []string{"3.3.3.3"})
	require.True(t, isWildcard)
	require.Contains(t, wildcards, "2.2.2.2")
	require.Equal(t, 3, hostLookups)
}

func TestResolverLookupHostIgnoresProbeErrors(t *testing.T) {
	resolver := NewResolver([]string{"example.com"}, func(host string) ([]string, error) {
		return nil, errors.New("lookup failed")
	})

	isWildcard, wildcards := resolver.LookupHost("www.example.com", []string{"1.1.1.1"})
	require.False(t, isWildcard)
	require.Empty(t, wildcards)
}

func TestResolverLookupHostDoesNotCacheProbeErrorsAsNormal(t *testing.T) {
	probeCalls := 0
	resolver := NewResolver([]string{"example.com"}, func(host string) ([]string, error) {
		switch host {
		case "target.example.com", "api.example.com":
			return []string{"1.1.1.1"}, nil
		default:
			if strings.HasSuffix(host, ".example.com") {
				probeCalls++
				if probeCalls == 1 {
					return nil, errors.New("temporary failure")
				}
				return []string{"1.1.1.1"}, nil
			}
			return nil, nil
		}
	})

	isWildcard, wildcards := resolver.LookupHost("target.example.com", []string{"1.1.1.1"})
	require.False(t, isWildcard)
	require.Empty(t, wildcards)

	isWildcard, wildcards = resolver.LookupHost("api.example.com", []string{"1.1.1.1"})
	require.True(t, isWildcard)
	require.Equal(t, map[string]struct{}{"1.1.1.1": {}}, wildcards)
}

func TestRegistrableRoot(t *testing.T) {
	tests := []struct {
		name string
		host string
		root string
		ok   bool
	}{
		{name: "fqdn", host: "WWW.Example.COM.", root: "example.com", ok: true},
		{name: "multi level suffix", host: "Api.Foo.Co.Uk.", root: "foo.co.uk", ok: true},
		{name: "wildcard input", host: "*.sub.example.com", root: "example.com", ok: true},
		{name: "ip address", host: net.ParseIP("127.0.0.1").String(), ok: false},
		{name: "invalid host", host: "localhost", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, ok := RegistrableRoot(tt.host)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.root, root)
		})
	}
}

func TestNewResolverWithDomainsSharesDomainSlice(t *testing.T) {
	domains := sliceutil.NewSyncSlice[string]()
	domains.Append("example.com")
	resolver := NewResolverWithDomains(domains, func(host string) ([]string, error) {
		return []string{"1.1.1.1"}, nil
	})

	domains.Append("example.org")
	isWildcard, _ := resolver.LookupHost("www.example.org", []string{"1.1.1.1"})
	require.True(t, isWildcard)
}
