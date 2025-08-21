// ABOUTME: Tests for lead image URL cleaner ensuring 100% JavaScript compatibility
// ABOUTME: Ports all test cases from the original JavaScript implementation

package cleaners

import (
	"testing"
)

func TestCleanLeadImageURLValidated_ValidURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "returns the url if valid",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "valid http url",
			input:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "valid https url with path",
			input:    "https://example.com/path/to/image.jpg",
			expected: "https://example.com/path/to/image.jpg",
		},
		{
			name:     "valid url with query parameters",
			input:    "https://example.com/image.jpg?w=800&h=600",
			expected: "https://example.com/image.jpg?w=800&h=600",
		},
		{
			name:     "valid url with fragment",
			input:    "https://example.com/image.jpg#section",
			expected: "https://example.com/image.jpg#section",
		},
		{
			name:     "valid url with port",
			input:    "https://example.com:8080/image.jpg",
			expected: "https://example.com:8080/image.jpg",
		},
		{
			name:     "localhost for development",
			input:    "http://localhost:3000/image.jpg",
			expected: "http://localhost:3000/image.jpg",
		},
		{
			name:     "IP address",
			input:    "http://192.168.1.1/image.jpg",
			expected: "http://192.168.1.1/image.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanLeadImageURLValidated(tt.input)
			if result == nil {
				t.Errorf("CleanLeadImageURLValidated(%q) = nil, expected %q", tt.input, tt.expected)
				return
			}
			if *result != tt.expected {
				t.Errorf("CleanLeadImageURLValidated(%q) = %q, expected %q", tt.input, *result, tt.expected)
			}
		})
	}
}

func TestCleanLeadImageURLValidated_InvalidURLs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "returns nil if the url is not valid",
			input: "this is not a valid url",
		},
		{
			name:  "invalid protocol",
			input: "ftp://example.com/image.jpg",
		},
		{
			name:  "javascript protocol (security)",
			input: "javascript:alert('xss')",
		},
		{
			name:  "data url",
			input: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
		},
		{
			name:  "file protocol",
			input: "file:///path/to/image.jpg",
		},
		{
			name:  "no protocol",
			input: "example.com/image.jpg",
		},
		{
			name:  "invalid host",
			input: "http:///image.jpg",
		},
		{
			name:  "malformed url",
			input: "http://",
		},
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "only whitespace",
			input: "   ",
		},
		{
			name:  "invalid characters",
			input: "https://exa mple.com/image.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanLeadImageURLValidated(tt.input)
			if result != nil {
				t.Errorf("CleanLeadImageURLValidated(%q) = %q, expected nil", tt.input, *result)
			}
		})
	}
}

func TestCleanLeadImageURLValidated_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trims leading whitespace",
			input:    "  https://example.com/foo/bar.jpg",
			expected: "https://example.com/foo/bar.jpg",
		},
		{
			name:     "trims trailing whitespace",
			input:    "https://example.com/foo/bar.jpg  ",
			expected: "https://example.com/foo/bar.jpg",
		},
		{
			name:     "trims both leading and trailing whitespace",
			input:    "  https://example.com/foo/bar.jpg  ",
			expected: "https://example.com/foo/bar.jpg",
		},
		{
			name:     "trims tabs and newlines",
			input:    "\t\nhttps://example.com/image.jpg\t\n",
			expected: "https://example.com/image.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanLeadImageURLValidated(tt.input)
			if result == nil {
				t.Errorf("CleanLeadImageURLValidated(%q) = nil, expected %q", tt.input, tt.expected)
				return
			}
			if *result != tt.expected {
				t.Errorf("CleanLeadImageURLValidated(%q) = %q, expected %q", tt.input, *result, tt.expected)
			}
		})
	}
}

func TestCleanLeadImageURLValidated_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{
			name:     "IPv6 address",
			input:    "http://[::1]:8080/image.jpg",
			expected: stringPtrLeadImage("http://[::1]:8080/image.jpg"),
		},
		{
			name:     "URL with authentication",
			input:    "https://user:pass@example.com/image.jpg",
			expected: stringPtrLeadImage("https://user:pass@example.com/image.jpg"),
		},
		{
			name:     "URL with international domain",
			input:    "https://例え.テスト/image.jpg",
			expected: stringPtrLeadImage("https://例え.テスト/image.jpg"),
		},
		{
			name:     "single word without dot (should be invalid)",
			input:    "http://example",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanLeadImageURLValidated(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("CleanLeadImageURLValidated(%q) = %q, expected nil", tt.input, *result)
				}
			} else {
				if result == nil {
					t.Errorf("CleanLeadImageURLValidated(%q) = nil, expected %q", tt.input, *tt.expected)
					return
				}
				if *result != *tt.expected {
					t.Errorf("CleanLeadImageURLValidated(%q) = %q, expected %q", tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

func TestCleanLeadImageURLValidatedString_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid url returns cleaned string",
			input:    "https://example.com/image.jpg",
			expected: "https://example.com/image.jpg",
		},
		{
			name:     "invalid url returns empty string",
			input:    "not a valid url",
			expected: "",
		},
		{
			name:     "empty input returns empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanLeadImageURLString(tt.input)
			if result != tt.expected {
				t.Errorf("CleanLeadImageURLValidatedString(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// stringPtrLeadImage returns a pointer to a string (helper for tests)
func stringPtrLeadImage(s string) *string {
	return &s
}

// BenchmarkCleanLeadImageURLValidated benchmarks the URL cleaning function
func BenchmarkCleanLeadImageURLValidated(b *testing.B) {
	url := "https://example.com/path/to/image.jpg"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CleanLeadImageURLValidated(url)
	}
}

// BenchmarkCleanLeadImageURLValidated_Invalid benchmarks with invalid URLs
func BenchmarkCleanLeadImageURLValidated_Invalid(b *testing.B) {
	url := "not a valid url at all"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CleanLeadImageURLValidated(url)
	}
}