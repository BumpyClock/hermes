// ABOUTME: Simple cleaners for various extracted fields when full cleaning logic is not available
// ABOUTME: Provides basic string cleaning and normalization for author, image URLs, and other fields

package cleaners

import (
	"strings"
	"net/url"
	
	"github.com/BumpyClock/hermes/internal/utils/text"
)

// CleanLeadImageURL ensures image URLs are properly formatted and absolute
func CleanLeadImageURL(imageURL, baseURL string) string {
	cleaned := strings.TrimSpace(imageURL)
	if cleaned == "" {
		return ""
	}
	
	// If URL is already absolute, return it
	if strings.HasPrefix(cleaned, "http://") || strings.HasPrefix(cleaned, "https://") {
		return cleaned
	}
	
	// If URL is protocol-relative, add https
	if strings.HasPrefix(cleaned, "//") {
		return "https:" + cleaned
	}
	
	// If URL is relative, make it absolute using baseURL
	if baseURL != "" {
		if base, err := url.Parse(baseURL); err == nil {
			if resolved, err := base.Parse(cleaned); err == nil {
				return resolved.String()
			}
		}
	}
	
	return cleaned
}

// CleanTitleSimple provides a simple title cleaner that doesn't need a document
func CleanTitleSimple(title, targetURL string) string {
	cleaned := strings.TrimSpace(title)
	if cleaned == "" {
		return ""
	}
	
	// Basic normalization
	cleaned = text.NormalizeSpaces(cleaned)
	
	// If title is too long (likely includes site name), try to shorten it
	if len(cleaned) > 150 {
		// Split on common separators and take the longest part
		separators := []string{" | ", " - ", ": "}
		for _, sep := range separators {
			if strings.Contains(cleaned, sep) {
				parts := strings.Split(cleaned, sep)
				longest := ""
				for _, part := range parts {
					if len(strings.TrimSpace(part)) > len(longest) {
						longest = strings.TrimSpace(part)
					}
				}
				if len(longest) > 10 && len(longest) < len(cleaned) {
					cleaned = longest
					break
				}
			}
		}
	}
	
	return cleaned
}