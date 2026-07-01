// ABOUTME: Simple cleaners for various extracted fields when full cleaning logic is not available
// ABOUTME: Provides basic string cleaning and normalization for author, image URLs, and other fields

package cleaners

import (
	"net/url"
	"strings"
)

// CleanLeadImageURL ensures image URLs are properly formatted and absolute.
func CleanLeadImageURL(imageURL, baseURL string) string {
	cleaned := strings.TrimSpace(imageURL)
	if cleaned == "" {
		return ""
	}

	resolved := cleaned
	if strings.HasPrefix(resolved, "//") {
		resolved = "https:" + resolved
	} else if !strings.HasPrefix(resolved, "http://") && !strings.HasPrefix(resolved, "https://") && baseURL != "" {
		if base, err := url.Parse(baseURL); err == nil {
			if absolute, err := base.Parse(resolved); err == nil {
				resolved = absolute.String()
			}
		}
	}

	if validated := CleanLeadImageURLValidated(resolved); validated != nil {
		return *validated
	}
	return ""
}
