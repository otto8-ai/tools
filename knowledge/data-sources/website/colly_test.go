package main

import "testing"

func TestIsSameDomainOrSubdomain(t *testing.T) {
	tests := []struct {
		linkHostname string
		baseHostname string
		expected     bool
	}{
		{"example.com", "example.com", true},      // exact match
		{"www.example.com", "example.com", true},  // www prefix in link
		{"www1.example.com", "example.com", true}, // www1 prefix in link
		{"sub.example.com", "example.com", false}, // not allowed, unrelated subdomain
		{"www.sub.example.com", "example.com", false},
		{"example.com", "www.example.com", true},  // link without www, base with www
		{"example.com", "www1.example.com", true}, // link without www, base with www1
		{"example.com", "www1.cn.example.com", false},
		{"example.com", "example.net", false},          // different base domains
		{"blog.example.com", "www.example.com", false}, // base with www, unrelated subdomain
		{"www.example.com", "www.example.com", true},   // exact match with www prefix
		{"www.ukcry.org", "cry.org", false},            // exact match with www prefix
	}

	for _, test := range tests {
		result := isSameDomainOrSubdomain(test.linkHostname, test.baseHostname)
		if result != test.expected {
			t.Errorf("For linkHostname: %s, baseHostname: %s - Expected: %v, Got: %v",
				test.linkHostname, test.baseHostname, test.expected, result)
		}
	}
}
