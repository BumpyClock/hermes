// ABOUTME: Custom error types for better error handling and debugging
// ABOUTME: Provides context-rich errors with URL, phase, and cause information

package parser

import (
	"fmt"
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


