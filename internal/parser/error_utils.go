// ABOUTME: Centralized error classification utilities to replace string matching with type-based error handling
// ABOUTME: Provides consistent error categorization using proper Go error handling patterns

package parser

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
)

// These constants mirror the public ErrorCode values
// We use int here to avoid import cycles - the caller will convert to their ErrorCode type
const (
	errInvalidURL = 0 // ErrInvalidURL
	errFetch      = 1 // ErrFetch
	errTimeout    = 2 // ErrTimeout
	errSSRF       = 3 // ErrSSRF
	errExtract    = 4 // ErrExtract
	errContext    = 5 // ErrContext (not used internally but keeps constants aligned)
)

// ClassifyErrorCode determines the appropriate error code based on the error type and context
// This replaces string-based error classification with proper type checking
// Returns an int that corresponds to the public ErrorCode values
func ClassifyErrorCode(err error, ctx context.Context, op string) int {
	if err == nil {
		return errFetch // Default fallback, shouldn't happen
	}
	
	// Check for context errors first (timeout/cancellation)
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return errTimeout
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return errTimeout // Treat cancellation as timeout for external API
		}
	}
	
	// Check for URL parsing errors
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		// url.Error can wrap various types of errors
		if isNetworkError(urlErr.Err) {
			return errFetch
		}
		if isTimeoutError(urlErr.Err) {
			return errTimeout
		}
		if isSSRFError(urlErr.Err) {
			return errSSRF
		}
		return errInvalidURL
	}
	
	// Check for network errors directly
	if isNetworkError(err) {
		return errFetch
	}
	
	// Check for timeout errors
	if isTimeoutError(err) {
		return errTimeout
	}
	
	// Check for SSRF protection errors
	if isSSRFError(err) {
		return errSSRF
	}
	
	// Check for extraction-specific errors by message patterns
	// This is less ideal but necessary for some internal errors
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "no children found") ||
		strings.Contains(errMsg, "failed to parse html") ||
		strings.Contains(errMsg, "content does not appear to be text") ||
		strings.Contains(errMsg, "document size") ||
		strings.Contains(errMsg, "dom too complex") {
		return errExtract
	}
	
	// Default to fetch error for unknown errors during HTTP operations
	return errFetch
}

// isNetworkError checks if an error is a network-related error
func isNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	
	// Check for specific network error types
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	
	return false
}

// isTimeoutError checks if an error represents a timeout
func isTimeoutError(err error) bool {
	// Check for timeout interface
	type timeout interface {
		Timeout() bool
	}
	
	if t, ok := err.(timeout); ok && t.Timeout() {
		return true
	}
	
	// Check for specific timeout errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	
	// Check for context timeout
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	
	return false
}

// isSSRFError checks if an error is related to SSRF protection
func isSSRFError(err error) bool {
	if err == nil {
		return false
	}
	
	errMsg := strings.ToLower(err.Error())
	
	// Check for URL validation failed errors (these are SSRF related)
	if strings.Contains(errMsg, "url validation failed") {
		return strings.Contains(errMsg, "private network") ||
			strings.Contains(errMsg, "localhost") ||
			strings.Contains(errMsg, "blocked")
	}
	
	// Check for other SSRF-specific patterns
	return strings.Contains(errMsg, "url not allowed") ||
		strings.Contains(errMsg, "ssrf")
}

