package hermes

import (
	"fmt"
)

// ErrorCode represents the type of error that occurred during parsing
type ErrorCode int

const (
	// ErrInvalidURL indicates the provided URL is malformed or empty
	ErrInvalidURL ErrorCode = iota
	
	// ErrFetch indicates a failure to fetch the content from the URL
	ErrFetch
	
	// ErrTimeout indicates the operation timed out
	ErrTimeout
	
	// ErrSSRF indicates the URL was blocked by SSRF protection
	ErrSSRF
	
	// ErrExtract indicates a failure during content extraction
	ErrExtract
	
	// ErrContext indicates the context was cancelled
	ErrContext
)

// String returns a human-readable string for the error code
func (e ErrorCode) String() string {
	switch e {
	case ErrInvalidURL:
		return "invalid URL"
	case ErrFetch:
		return "fetch error"
	case ErrTimeout:
		return "timeout"
	case ErrSSRF:
		return "SSRF blocked"
	case ErrExtract:
		return "extraction error"
	case ErrContext:
		return "context cancelled"
	default:
		return "unknown error"
	}
}

// ParseError represents an error that occurred during parsing.
// It includes the error code, URL, operation, and underlying error.
type ParseError struct {
	// Code indicates the type of error
	Code ErrorCode
	
	// URL is the URL that was being parsed when the error occurred
	URL string
	
	// Op is the operation that failed (e.g., "Parse", "ParseHTML")
	Op string
	
	// Err is the underlying error
	Err error
}

// Error implements the error interface
func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("hermes: %s %s: %s: %v", e.Op, e.URL, e.Code, e.Err)
	}
	return fmt.Sprintf("hermes: %s %s: %s", e.Op, e.URL, e.Code)
}

// Unwrap returns the underlying error
func (e *ParseError) Unwrap() error {
	return e.Err
}

// Is reports whether the target error is equal to this error
func (e *ParseError) Is(target error) bool {
	t, ok := target.(*ParseError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// IsTimeout returns true if the error was caused by a timeout
func (e *ParseError) IsTimeout() bool {
	return e.Code == ErrTimeout
}

// IsSSRF returns true if the error was caused by SSRF protection
func (e *ParseError) IsSSRF() bool {
	return e.Code == ErrSSRF
}

// IsFetch returns true if the error occurred during content fetching
func (e *ParseError) IsFetch() bool {
	return e.Code == ErrFetch
}

// IsExtract returns true if the error occurred during content extraction
func (e *ParseError) IsExtract() bool {
	return e.Code == ErrExtract
}

// IsInvalidURL returns true if the error was caused by an invalid URL
func (e *ParseError) IsInvalidURL() bool {
	return e.Code == ErrInvalidURL
}

// IsContext returns true if the error was caused by context cancellation
func (e *ParseError) IsContext() bool {
	return e.Code == ErrContext
}