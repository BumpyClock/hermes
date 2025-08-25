// ABOUTME: Custom error types for better error handling and debugging
// ABOUTME: Provides context-rich errors with URL, phase, and cause information

package parser

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ParseError represents an error that occurred during parsing
type ParseError struct {
	URL       string    `json:"url"`                 // URL being parsed when error occurred
	Phase     string    `json:"phase"`               // Parse phase: "fetch", "extract", "clean", etc.
	Err       error     `json:"error"`               // Underlying error
	Timestamp time.Time `json:"timestamp"`           // When the error occurred
	Field     string    `json:"field,omitempty"`     // Specific field being extracted (if applicable)
	Selector  string    `json:"selector,omitempty"`  // CSS selector being processed (if applicable)
	Message   string    `json:"message,omitempty"`   // Additional context message
}

// Error implements the error interface
func (pe *ParseError) Error() string {
	var parts []string
	
	if pe.Phase != "" {
		parts = append(parts, fmt.Sprintf("phase:%s", pe.Phase))
	}
	
	if pe.URL != "" {
		parts = append(parts, fmt.Sprintf("url:%s", pe.URL))
	}
	
	if pe.Field != "" {
		parts = append(parts, fmt.Sprintf("field:%s", pe.Field))
	}
	
	if pe.Selector != "" {
		parts = append(parts, fmt.Sprintf("selector:%s", pe.Selector))
	}
	
	if pe.Message != "" {
		parts = append(parts, pe.Message)
	}
	
	if pe.Err != nil {
		parts = append(parts, pe.Err.Error())
	}
	
	return strings.Join(parts, " | ")
}

// Unwrap returns the underlying error for error unwrapping
func (pe *ParseError) Unwrap() error {
	return pe.Err
}

// Is supports error checking with errors.Is()
func (pe *ParseError) Is(target error) bool {
	if target == nil {
		return false
	}
	
	if otherPE, ok := target.(*ParseError); ok {
		return pe.Phase == otherPE.Phase && pe.URL == otherPE.URL
	}
	
	return pe.Err != nil && pe.Err.Error() == target.Error()
}

// ParseErrorType represents different categories of parse errors
type ParseErrorType string

const (
	ErrorTypeFetch     ParseErrorType = "fetch"      // Network/HTTP errors
	ErrorTypeExtract   ParseErrorType = "extract"    // Content extraction errors
	ErrorTypeClean     ParseErrorType = "clean"      // Content cleaning errors
	ErrorTypeValidate  ParseErrorType = "validate"   // Input validation errors
	ErrorTypeTransform ParseErrorType = "transform"  // Content transformation errors
	ErrorTypeTimeout   ParseErrorType = "timeout"    // Timeout errors
	ErrorTypeResource  ParseErrorType = "resource"   // Resource loading errors
)

// NewParseError creates a new ParseError with context
func NewParseError(phase string, url string, err error) *ParseError {
	return &ParseError{
		URL:       url,
		Phase:     phase,
		Err:       err,
		Timestamp: time.Now(),
	}
}

// NewFetchError creates an error for HTTP/network issues
func NewFetchError(url string, err error) *ParseError {
	return &ParseError{
		URL:       url,
		Phase:     string(ErrorTypeFetch),
		Err:       err,
		Timestamp: time.Now(),
	}
}

// NewExtractionError creates an error for content extraction issues
func NewExtractionError(url string, field string, selector string, err error) *ParseError {
	return &ParseError{
		URL:       url,
		Phase:     string(ErrorTypeExtract),
		Field:     field,
		Selector:  selector,
		Err:       err,
		Timestamp: time.Now(),
	}
}

// NewValidationError creates an error for input validation issues
func NewValidationError(url string, message string, err error) *ParseError {
	return &ParseError{
		URL:       url,
		Phase:     string(ErrorTypeValidate),
		Message:   message,
		Err:       err,
		Timestamp: time.Now(),
	}
}

// NewTimeoutError creates an error for timeout issues
func NewTimeoutError(url string, phase string, duration time.Duration) *ParseError {
	return &ParseError{
		URL:       url,
		Phase:     phase,
		Message:   fmt.Sprintf("timeout after %v", duration),
		Err:       fmt.Errorf("operation timed out"),
		Timestamp: time.Now(),
	}
}

// WithField adds field context to an existing error
func (pe *ParseError) WithField(field string) *ParseError {
	pe.Field = field
	return pe
}

// WithSelector adds selector context to an existing error
func (pe *ParseError) WithSelector(selector string) *ParseError {
	pe.Selector = selector
	return pe
}

// WithMessage adds additional context message
func (pe *ParseError) WithMessage(message string) *ParseError {
	pe.Message = message
	return pe
}

// IsNetworkError checks if the error is network-related
func (pe *ParseError) IsNetworkError() bool {
	return pe.Phase == string(ErrorTypeFetch)
}

// IsExtractionError checks if the error is extraction-related
func (pe *ParseError) IsExtractionError() bool {
	return pe.Phase == string(ErrorTypeExtract)
}

// IsValidationError checks if the error is validation-related
func (pe *ParseError) IsValidationError() bool {
	return pe.Phase == string(ErrorTypeValidate)
}

// IsTimeoutError checks if the error is timeout-related
func (pe *ParseError) IsTimeoutError() bool {
	return pe.Phase == string(ErrorTypeTimeout) || 
		   (pe.Message != "" && strings.Contains(pe.Message, "timeout"))
}

// GetDomain extracts the domain from the URL
func (pe *ParseError) GetDomain() string {
	if pe.URL == "" {
		return ""
	}
	
	if parsedURL, err := url.Parse(pe.URL); err == nil {
		return parsedURL.Host
	}
	
	return ""
}

// ErrorCollection holds multiple parse errors
type ErrorCollection struct {
	Errors []*ParseError `json:"errors"`
}

// Add adds a new error to the collection
func (ec *ErrorCollection) Add(err *ParseError) {
	ec.Errors = append(ec.Errors, err)
}

// HasErrors returns true if there are any errors
func (ec *ErrorCollection) HasErrors() bool {
	return len(ec.Errors) > 0
}

// Count returns the number of errors
func (ec *ErrorCollection) Count() int {
	return len(ec.Errors)
}

// Error implements the error interface for ErrorCollection
func (ec *ErrorCollection) Error() string {
	if len(ec.Errors) == 0 {
		return "no errors"
	}
	
	if len(ec.Errors) == 1 {
		return ec.Errors[0].Error()
	}
	
	var parts []string
	for i, err := range ec.Errors {
		parts = append(parts, fmt.Sprintf("[%d] %s", i+1, err.Error()))
	}
	
	return fmt.Sprintf("multiple errors: %s", strings.Join(parts, "; "))
}

// GetByPhase returns all errors from a specific phase
func (ec *ErrorCollection) GetByPhase(phase string) []*ParseError {
	var result []*ParseError
	for _, err := range ec.Errors {
		if err.Phase == phase {
			result = append(result, err)
		}
	}
	return result
}

// GetByURL returns all errors from a specific URL
func (ec *ErrorCollection) GetByURL(url string) []*ParseError {
	var result []*ParseError
	for _, err := range ec.Errors {
		if err.URL == url {
			result = append(result, err)
		}
	}
	return result
}

// HasPhaseErrors checks if there are errors in a specific phase
func (ec *ErrorCollection) HasPhaseErrors(phase string) bool {
	return len(ec.GetByPhase(phase)) > 0
}

// First returns the first error or nil if no errors
func (ec *ErrorCollection) First() *ParseError {
	if len(ec.Errors) > 0 {
		return ec.Errors[0]
	}
	return nil
}

// Last returns the last error or nil if no errors
func (ec *ErrorCollection) Last() *ParseError {
	if len(ec.Errors) > 0 {
		return ec.Errors[len(ec.Errors)-1]
	}
	return nil
}

// Clear removes all errors from the collection
func (ec *ErrorCollection) Clear() {
	ec.Errors = ec.Errors[:0]
}

// WrapError wraps a regular error as a ParseError if it isn't already one
func WrapError(err error, phase string, url string) error {
	if err == nil {
		return nil
	}
	
	if pe, ok := err.(*ParseError); ok {
		// Already a ParseError, just update context if missing
		if pe.Phase == "" {
			pe.Phase = phase
		}
		if pe.URL == "" {
			pe.URL = url
		}
		return pe
	}
	
	return NewParseError(phase, url, err)
}

// ConvertError converts any error to a ParseError for consistent error handling
func ConvertError(err error) *ParseError {
	if err == nil {
		return nil
	}
	
	if pe, ok := err.(*ParseError); ok {
		return pe
	}
	
	return &ParseError{
		Phase:     "unknown",
		Err:       err,
		Timestamp: time.Now(),
	}
}