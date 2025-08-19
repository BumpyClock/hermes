// ABOUTME: Tests for RemoveAnchor function ensuring JavaScript compatibility
// ABOUTME: Covers all edge cases including fragment formats and URL variations
package text

import (
	"testing"
)

func TestRemoveAnchor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "returns URL without anchor from JavaScript test case 1",
			input:    "http://example.com/foo/bar/wow-cool/page=10/#wow",
			expected: "http://example.com/foo/bar/wow-cool/page=10",
		},
		{
			name:     "returns same URL if no anchor found from JavaScript test case 2",
			input:    "http://example.com/foo/bar/wow-cool",
			expected: "http://example.com/foo/bar/wow-cool",
		},
		{
			name:     "removes anchor with complex fragment",
			input:    "https://site.com/article#section-1-detailed",
			expected: "https://site.com/article",
		},
		{
			name:     "removes anchor and trailing slash",
			input:    "https://example.com/path/#anchor",
			expected: "https://example.com/path",
		},
		{
			name:     "removes only trailing slash when no anchor",
			input:    "https://example.com/path/",
			expected: "https://example.com/path",
		},
		{
			name:     "handles multiple hash characters (only first is fragment)",
			input:    "https://example.com/path#anchor#extra",
			expected: "https://example.com/path",
		},
		{
			name:     "handles empty anchor",
			input:    "https://example.com/path#",
			expected: "https://example.com/path",
		},
		{
			name:     "handles root URL with anchor",
			input:    "https://example.com/#home",
			expected: "https://example.com",
		},
		{
			name:     "handles root URL with trailing slash only",
			input:    "https://example.com/",
			expected: "https://example.com",
		},
		{
			name:     "handles query parameters with anchor",
			input:    "https://example.com/search?q=test&page=1#results",
			expected: "https://example.com/search?q=test&page=1",
		},
		{
			name:     "handles anchor with URL-encoded characters",
			input:    "https://example.com/page#section%20name",
			expected: "https://example.com/page",
		},
		{
			name:     "preserves query parameters without anchor",
			input:    "https://example.com/search?q=test&page=1",
			expected: "https://example.com/search?q=test&page=1",
		},
		{
			name:     "handles relative URLs with anchor",
			input:    "/path/to/page#section",
			expected: "/path/to/page",
		},
		{
			name:     "handles protocol-relative URLs with anchor",
			input:    "//example.com/path#section",
			expected: "//example.com/path",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles URL with port and anchor",
			input:    "http://localhost:3000/app#dashboard",
			expected: "http://localhost:3000/app",
		},
		{
			name:     "handles deeply nested path with anchor",
			input:    "https://example.com/very/deep/path/structure/page.html#conclusion",
			expected: "https://example.com/very/deep/path/structure/page.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveAnchor(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveAnchor(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// BenchmarkRemoveAnchor measures performance of the function
func BenchmarkRemoveAnchor(b *testing.B) {
	testURL := "https://example.com/very/long/path/with/many/segments/and/parameters?param1=value1&param2=value2#very-long-anchor-name-with-dashes"
	
	for i := 0; i < b.N; i++ {
		RemoveAnchor(testURL)
	}
}

// TestRemoveAnchorJavaScriptCompatibility ensures exact JavaScript behavior
func TestRemoveAnchorJavaScriptCompatibility(t *testing.T) {
	// Test the exact behavior from the JavaScript tests
	// These test cases are directly from remove-anchor.test.js
	
	t.Run("JavaScript test case 1 - URL with anchor and trailing slash", func(t *testing.T) {
		url := "http://example.com/foo/bar/wow-cool/page=10/#wow"
		expected := "http://example.com/foo/bar/wow-cool/page=10"
		result := RemoveAnchor(url)
		
		if result != expected {
			t.Errorf("RemoveAnchor(%q) = %q, expected %q (JavaScript compatibility test 1)", url, result, expected)
		}
	})
	
	t.Run("JavaScript test case 2 - URL without anchor", func(t *testing.T) {
		url := "http://example.com/foo/bar/wow-cool"
		expected := "http://example.com/foo/bar/wow-cool"
		result := RemoveAnchor(url)
		
		if result != expected {
			t.Errorf("RemoveAnchor(%q) = %q, expected %q (JavaScript compatibility test 2)", url, result, expected)
		}
	})
}