// ABOUTME: URL validation utilities for preventing injection attacks and malicious URLs
// ABOUTME: Provides comprehensive validation against path traversal, SSRF, and other URL-based attacks

package security

import (
	"context"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	// Allowed URL schemes for web content
	allowedSchemes = map[string]bool{
		"http":  true,
		"https": true,
	}
	
	// Dangerous URL patterns to reject
	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\.\./`),                    // Path traversal
		regexp.MustCompile(`%2e%2e%2f`),               // URL-encoded path traversal
		regexp.MustCompile(`javascript:`),             // JavaScript protocol
		regexp.MustCompile(`data:`),                   // Data URLs
		regexp.MustCompile(`file:`),                   // File protocol
		regexp.MustCompile(`ftp:`),                    // FTP protocol
		regexp.MustCompile(`\x00`),                    // Null bytes
		regexp.MustCompile(`[\x01-\x08\x0B\x0C\x0E-\x1F\x7F]`), // Control characters
	}
	
	// Private/internal IP ranges to block (SSRF prevention)
	privateNetworks = []string{
		"127.0.0.0/8",    // Loopback
		"10.0.0.0/8",     // Private A
		"172.16.0.0/12",  // Private B
		"192.168.0.0/16", // Private C
		"169.254.0.0/16", // Link-local
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 private
		"fe80::/10",      // IPv6 link-local
	}
	
	// Compiled private networks for efficiency
	privateIPNets []*net.IPNet
)

func init() {
	// Pre-compile private network ranges
	for _, cidr := range privateNetworks {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			privateIPNets = append(privateIPNets, ipNet)
		}
	}
}

// ValidateURL performs comprehensive URL validation for security
// DEPRECATED: This method uses context.Background() which prevents proper timeout control.
// Use ValidateURLWithContext instead.
func ValidateURL(rawURL string) error {
	return ValidateURLWithContext(context.Background(), rawURL)
}

// ValidateURLWithContext performs comprehensive URL validation for security with context support
func ValidateURLWithContext(ctx context.Context, rawURL string) error {
	// Use default options (private networks not allowed)
	return ValidateURLWithOptions(ctx, rawURL, false)
}

// ValidateURLWithOptions performs URL validation with configurable options
func ValidateURLWithOptions(ctx context.Context, rawURL string, allowPrivateNetworks bool) error {
	if rawURL == "" {
		return &URLValidationError{Type: "empty", Message: "URL cannot be empty"}
	}
	
	// Check for dangerous patterns in raw URL
	lowerURL := strings.ToLower(rawURL)
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(lowerURL) {
			return &URLValidationError{Type: "dangerous_pattern", Message: "URL contains dangerous patterns"}
		}
	}
	
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return &URLValidationError{Type: "malformed", Message: "URL is malformed: " + err.Error()}
	}
	
	// Validate scheme
	if !allowedSchemes[strings.ToLower(parsedURL.Scheme)] {
		return &URLValidationError{Type: "invalid_scheme", Message: "URL scheme not allowed: " + parsedURL.Scheme}
	}
	
	// Validate host presence
	if parsedURL.Host == "" {
		return &URLValidationError{Type: "no_host", Message: "URL must have a host"}
	}
	
	// Check for private/internal IPs (SSRF prevention) if not explicitly allowed
	if !allowPrivateNetworks {
		if err := validateHostNotPrivateWithContext(ctx, parsedURL.Host); err != nil {
			return err
		}
	}
	
	// Additional path validation
	if strings.Contains(parsedURL.Path, "..") {
		return &URLValidationError{Type: "path_traversal", Message: "Path traversal detected in URL path"}
	}
	
	return nil
}

// validateHostNotPrivate checks if the host resolves to private IP addresses
// DEPRECATED: This method uses context.Background() which prevents proper timeout control.
// Use validateHostNotPrivateWithContext instead.
func validateHostNotPrivate(host string) error {
	return validateHostNotPrivateWithContext(context.Background(), host)
}

// validateHostNotPrivateWithContext checks if the host resolves to private IP addresses with context support
func validateHostNotPrivateWithContext(ctx context.Context, host string) error {
	// Extract hostname from host:port format
	hostname := host
	if strings.Contains(host, ":") {
		var err error
		hostname, _, err = net.SplitHostPort(host)
		if err != nil {
			return &URLValidationError{Type: "invalid_host", Message: "Invalid host format: " + err.Error()}
		}
	}
	
	// Skip DNS resolution for IP addresses
	if ip := net.ParseIP(hostname); ip != nil {
		return validateIPNotPrivate(ip)
	}
	
	// For domain names, resolve and check all IPs using context-aware resolver
	resolver := &net.Resolver{}
	ips, err := resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			return &URLValidationError{Type: "context_cancelled", Message: "DNS resolution cancelled: " + ctx.Err().Error()}
		}
		return &URLValidationError{Type: "dns_error", Message: "Failed to resolve hostname: " + err.Error()}
	}
	
	for _, ipAddr := range ips {
		if err := validateIPNotPrivate(ipAddr.IP); err != nil {
			return err
		}
	}
	
	return nil
}

// validateIPNotPrivate checks if an IP is in private/internal ranges
func validateIPNotPrivate(ip net.IP) error {
	for _, ipNet := range privateIPNets {
		if ipNet.Contains(ip) {
			return &URLValidationError{Type: "private_ip", Message: "URL resolves to private/internal IP address: " + ip.String()}
		}
	}
	return nil
}

// IsValidWebURL performs basic validation for web URLs (less strict than ValidateURL)
func IsValidWebURL(u *url.URL) bool {
	return u != nil && 
		   (u.Scheme == "http" || u.Scheme == "https") && 
		   u.Host != ""
}

// URLValidationError represents URL validation errors
type URLValidationError struct {
	Type    string
	Message string
}

func (e *URLValidationError) Error() string {
	return e.Message
}

// IsSSRFError checks if the error is related to SSRF protection
func IsSSRFError(err error) bool {
	if urlErr, ok := err.(*URLValidationError); ok {
		return urlErr.Type == "private_ip" || urlErr.Type == "dns_error"
	}
	return false
}