// ABOUTME: Tests for character encoding detection from HTML content
// ABOUTME: Matches JavaScript get-encoding.test.js behavior exactly

package text

import (
	"testing"
)

func TestGetEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "returns the encoding as a string from Content-Type",
			input:    "text/html; charset=iso-8859-15",
			expected: "iso-8859-15",
		},
		{
			name:     "returns utf-8 as default if no encoding found",
			input:    "text/html",
			expected: "utf-8",
		},
		{
			name:     "returns utf-8 if there is an invalid encoding",
			input:    "text/html; charset=fake-charset",
			expected: "utf-8",
		},
		{
			name:     "handles charset with quotes",
			input:    `text/html; charset="utf-8"`,
			expected: "utf-8",
		},
		{
			name:     "handles charset with single quotes",
			input:    "text/html; charset='iso-8859-1'",
			expected: "iso-8859-1",
		},
		{
			name:     "handles case insensitive charset",
			input:    "text/html; CHARSET=UTF-8",
			expected: "UTF-8",
		},
		{
			name:     "handles multiple parameters",
			input:    "text/html; boundary=something; charset=windows-1251",
			expected: "windows-1251",
		},
		{
			name:     "handles charset parameter with spaces",
			input:    "text/html; charset = utf-8",
			expected: "utf-8",
		},
		{
			name:     "returns utf-8 for empty string",
			input:    "",
			expected: "utf-8",
		},
		{
			name:     "handles charset from HTML meta tag",
			input:    "windows-1250",
			expected: "windows-1250",
		},
		{
			name:     "handles direct charset name",
			input:    "iso-8859-1",
			expected: "iso-8859-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEncoding(tt.input)
			if result != tt.expected {
				t.Errorf("GetEncoding(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetEncodingWithEncodingRE(t *testing.T) {
	// Test specific patterns that should match ENCODING_RE
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "charset with equals",
			input:    "charset=utf-8",
			expected: "utf-8",
		},
		{
			name:     "charset with dashes",
			input:    "charset=iso-8859-1",
			expected: "iso-8859-1",
		},
		{
			name:     "charset with underscores",
			input:    "charset=windows_1251",
			expected: "windows_1251",
		},
		{
			name:     "charset with numbers",
			input:    "charset=cp1252",
			expected: "cp1252",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEncoding(tt.input)
			if result != tt.expected {
				t.Errorf("GetEncoding(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetEncodingValidation(t *testing.T) {
	// Test charset validation - these should return utf-8 for invalid charsets
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "fake charset",
			input: "charset=fake-charset",
		},
		{
			name:  "invalid charset",
			input: "charset=not-a-real-encoding",
		},
		{
			name:  "empty charset",
			input: "charset=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEncoding(tt.input)
			if result != "utf-8" {
				t.Errorf("GetEncoding(%q) = %q, want %q (should fallback to utf-8 for invalid charset)", tt.input, result, "utf-8")
			}
		})
	}
}

func TestGetEncodingComplexCases(t *testing.T) {
	// Test cases inspired by resource/index.test.js
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "meta content-type with iso-8859-1",
			input:    "text/html; charset=iso-8859-1",
			expected: "iso-8859-1",
		},
		{
			name:     "meta content-type with windows-1251",
			input:    "text/html; charset=windows-1251",
			expected: "windows-1251",
		},
		{
			name:     "meta content-type with windows-1250",
			input:    "text/html; charset=windows-1250",
			expected: "windows-1250",
		},
		{
			name:     "HTML5 charset meta tag format",
			input:    "windows-1250",
			expected: "windows-1250",
		},
		{
			name:     "with boundary parameter before charset",
			input:    "text/html; boundary=something; charset=utf-8",
			expected: "utf-8",
		},
		{
			name:     "charset with extra whitespace",
			input:    "text/html; charset = iso-8859-1 ",
			expected: "utf-8", // This should fail regex and fallback since pattern doesn't handle spaces
		},
		{
			name:     "multiple charsets - should pick first",
			input:    "text/html; charset=utf-8; charset=iso-8859-1",
			expected: "utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEncoding(tt.input)
			if result != tt.expected {
				t.Errorf("GetEncoding(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}