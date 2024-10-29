package retrievers

import (
	"regexp"
	"strings"
)

// regex pattern to match double-quoted substrings
var doubleQuotePattern = regexp.MustCompile(`"([^"]*)"`)

// Extract double-quoted substrings from a string
func ExtractQuotedSubstrings(input string) []string {
	matches := doubleQuotePattern.FindAllStringSubmatch(input, -1)

	var substrings []string
	for _, match := range matches {
		if len(match) > 1 {
			m := strings.TrimSpace(match[1])
			if m != "" {
				substrings = append(substrings, m)
			}
		}
	}

	return substrings
}
