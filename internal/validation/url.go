// ABOUTME: Unified URL validation pipeline consolidating all validation logic into a single, consistent interface
// ABOUTME: Replaces scattered validation functions with a comprehensive, configurable validation system

package validation

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// ValidationOptions configures URL validation behavior
type ValidationOptions struct {
	AllowPrivateNetworks bool
	AllowLocalhost       bool
	RequireHTTPS        bool
	MaxHostnameLength   int
	Timeout             time.Duration
}

// DefaultValidationOptions returns secure defaults for URL validation
func DefaultValidationOptions() ValidationOptions {
	return ValidationOptions{
		AllowPrivateNetworks: false,
		AllowLocalhost:      false,
		RequireHTTPS:        false,
		MaxHostnameLength:   253, // RFC 1035 limit
		Timeout:             5 * time.Second,
	}
}

// ValidationError represents a URL validation error with specific type information
type ValidationError struct {
	Type    string
	Message string
	URL     string
}

func (e *ValidationError) Error() string {
	if e.URL != "" {
		return fmt.Sprintf("URL validation failed (%s): %s - %s", e.Type, e.Message, e.URL)
	}
	return fmt.Sprintf("URL validation failed (%s): %s", e.Type, e.Message)
}

// ValidateURL performs comprehensive URL validation with configurable options
// This is the main entry point that consolidates all validation logic
func ValidateURL(ctx context.Context, rawURL string, opts ValidationOptions) error {
	if rawURL == "" {
		return &ValidationError{
			Type:    "empty",
			Message: "URL cannot be empty",
			URL:     rawURL,
		}
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return &ValidationError{
			Type:    "parse",
			Message: fmt.Sprintf("failed to parse URL: %v", err),
			URL:     rawURL,
		}
	}

	// Basic structure validation
	if err := validateBasicStructure(parsedURL); err != nil {
		return err
	}

	// Hostname length validation
	if len(parsedURL.Host) > opts.MaxHostnameLength {
		return &ValidationError{
			Type:    "hostname_length",
			Message: fmt.Sprintf("hostname too long (%d chars, max %d)", len(parsedURL.Host), opts.MaxHostnameLength),
			URL:     rawURL,
		}
	}

	// HTTPS requirement check
	if opts.RequireHTTPS && parsedURL.Scheme != "https" {
		return &ValidationError{
			Type:    "scheme",
			Message: "HTTPS is required",
			URL:     rawURL,
		}
	}

	// Network-based validation (with context for timeout)
	if err := validateNetworkAccess(ctx, parsedURL, opts); err != nil {
		return err
	}

	return nil
}

// ValidateURLSimple performs basic URL validation without network checks
// Useful for quick validation where SSRF protection isn't needed
func ValidateURLSimple(rawURL string) error {
	if rawURL == "" {
		return &ValidationError{Type: "empty", Message: "URL cannot be empty", URL: rawURL}
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return &ValidationError{Type: "parse", Message: fmt.Sprintf("failed to parse: %v", err), URL: rawURL}
	}

	return validateBasicStructure(parsedURL)
}

// validateBasicStructure checks basic URL structure requirements
func validateBasicStructure(u *url.URL) error {
	if u.Scheme == "" {
		return &ValidationError{Type: "scheme", Message: "URL scheme is required", URL: u.String()}
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return &ValidationError{Type: "scheme", Message: "only HTTP and HTTPS schemes are allowed", URL: u.String()}
	}

	if u.Host == "" {
		return &ValidationError{Type: "host", Message: "URL host is required", URL: u.String()}
	}

	// Check for suspicious characters that could indicate injection
	if strings.ContainsAny(u.Host, " \t\n\r") {
		return &ValidationError{Type: "host", Message: "host contains invalid characters", URL: u.String()}
	}

	return nil
}

// validateNetworkAccess performs network-based validation including SSRF protection
func validateNetworkAccess(ctx context.Context, u *url.URL, opts ValidationOptions) error {
	// Extract hostname from host (remove port if present)
	hostname := u.Hostname()
	if hostname == "" {
		return &ValidationError{Type: "host", Message: "cannot extract hostname", URL: u.String()}
	}

	// Check localhost restrictions
	if !opts.AllowLocalhost && isLocalhost(hostname) {
		return &ValidationError{Type: "localhost", Message: "localhost access not allowed", URL: u.String()}
	}

	// DNS resolution with context timeout
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	resolver := &net.Resolver{}
	addrs, err := resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ValidationError{Type: "dns_timeout", Message: "DNS resolution timed out", URL: u.String()}
		}
		return &ValidationError{Type: "dns", Message: fmt.Sprintf("DNS resolution failed: %v", err), URL: u.String()}
	}

	if len(addrs) == 0 {
		return &ValidationError{Type: "dns", Message: "no IP addresses found", URL: u.String()}
	}

	// Check for private networks if not allowed
	if !opts.AllowPrivateNetworks {
		for _, addr := range addrs {
			if isPrivateIP(addr.IP) {
				return &ValidationError{Type: "private_network", Message: "private network access not allowed", URL: u.String()}
			}
		}
	}

	return nil
}

// isLocalhost checks if a hostname refers to localhost
func isLocalhost(hostname string) bool {
	return hostname == "localhost" || 
		   hostname == "127.0.0.1" || 
		   hostname == "::1" ||
		   strings.HasSuffix(hostname, ".localhost")
}

// isPrivateIP checks if an IP address is in a private network range
func isPrivateIP(ip net.IP) bool {
	// IPv4 private ranges
	private4 := []string{
		"10.0.0.0/8",     // RFC 1918
		"172.16.0.0/12",  // RFC 1918  
		"192.168.0.0/16", // RFC 1918
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local
	}

	// IPv6 private ranges
	private6 := []string{
		"::1/128",      // Loopback
		"fc00::/7",     // Unique local
		"fe80::/10",    // Link-local
	}

	allRanges := append(private4, private6...)
	
	for _, cidr := range allRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// IsValidWebURL performs lightweight validation for web URLs (backward compatibility)
func IsValidWebURL(u *url.URL) bool {
	return u != nil && 
		   (u.Scheme == "http" || u.Scheme == "https") && 
		   u.Host != ""
}