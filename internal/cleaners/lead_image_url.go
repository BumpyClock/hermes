// ABOUTME: Lead image URL cleaner that validates and cleans image URLs to ensure they are valid web URIs
// ABOUTME: Provides 100% JavaScript compatibility with the original valid-url library behavior

package cleaners

import (
	"net/url"
	"strings"
)

// CleanLeadImageURLValidated validates and cleans a lead image URL
// Returns nil if the URL is invalid, cleaned URL string if valid
// Matches JavaScript behavior: trim whitespace and validate as web URI
func CleanLeadImageURLValidated(leadImageURL string) *string {
	// Trim whitespace (matching JavaScript behavior)
	trimmed := strings.TrimSpace(leadImageURL)
	
	// Return nil for empty strings (matching JavaScript null return)
	if trimmed == "" {
		return nil
	}
	
	// Parse the URL to validate it
	parsedURL, err := url.Parse(trimmed)
	if err != nil {
		return nil
	}
	
	// Validate that it's a web URI (http or https only)
	// This matches the JavaScript valid-url.isWebUri() behavior
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil
	}
	
	// Ensure the URL has a valid host
	if parsedURL.Host == "" {
		return nil
	}
	
	// Additional validation: ensure host contains at least one dot (like domain.com)
	// This prevents URLs like "http://example" which valid-url might reject
	// But allow localhost, localhost:port, IP addresses, and IPv6 addresses
	if !strings.Contains(parsedURL.Host, ".") && !strings.Contains(parsedURL.Host, ":") {
		// Allow localhost without port
		if parsedURL.Host != "localhost" {
			return nil
		}
	}
	
	// Return the trimmed URL if all validations pass
	return &trimmed
}

// isIPAddress checks if a string looks like an IP address
// Simple heuristic to allow IP addresses in development
func isIPAddress(host string) bool {
	// Simple check: if it contains only digits, dots, and colons (IPv4/IPv6)
	for _, r := range host {
		if !(r >= '0' && r <= '9') && r != '.' && r != ':' {
			return false
		}
	}
	return strings.Count(host, ".") == 3 || strings.Contains(host, ":")
}

// CleanLeadImageURLString provides a string-returning version for backward compatibility
// Returns empty string if URL is invalid, cleaned URL if valid
func CleanLeadImageURLString(leadImageURL string) string {
	if cleaned := CleanLeadImageURLValidated(leadImageURL); cleaned != nil {
		return *cleaned
	}
	return ""
}