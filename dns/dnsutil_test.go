package dnsutil

import (
	"testing"
)

func TestDecomposeDomain(t *testing.T) {
	tests := []struct {
		name string
		subdomain  string
		domain     string
		expectErr  bool
	}{
		{"www.example.com", "www", "example.com", false},
		{"http://www.example.com", "www", "example.com", false},
		{"example.com", "", "example.com", false},
		{"sub.sub.example.co.uk", "sub.sub", "example.co.uk", false},
		{"invalid_domain", "", "", true},
		{"", "", "", true},
	}

	for _, test := range tests {
		subdomain, domain, err := DecomposeDomain(test.name)
		if test.expectErr && err == nil {
			t.Errorf("expected error for domain %s, but got none", test.name)
		}
		if !test.expectErr && err != nil {
			t.Errorf("did not expect error for domain %s, but got %v", test.name, err)
		}
		if subdomain != test.subdomain {
			t.Errorf("expected subdomain %s for domain %s, but got %s", test.subdomain, test.name, subdomain)
		}
		if domain != test.domain {
			t.Errorf("expected domain %s for domain %s, but got %s", test.domain, test.name, domain)
		}
	}
}
