// ABOUTME: Test file for PageNumFromURL function that extracts page numbers from URLs
// ABOUTME: Tests all URL patterns including page=N, pg=N, pagination/N, etc.

package text

import (
	"testing"
)

func TestPageNumFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected *int
	}{
		// Test cases that should return nil (no page number)
		{
			name:     "no page number in URL",
			url:      "http://example.com",
			expected: nil,
		},
		{
			name:     "page number too large (>= 100)",
			url:      "http://example.com/?pg=102",
			expected: nil,
		},
		{
			name:     "wrong separator (colon instead of equals)",
			url:      "http://example.com/?page:102",
			expected: nil,
		},
		// Test cases that should return page numbers
		{
			name:     "page parameter with equals",
			url:      "http://example.com/foo?page=1",
			expected: intPtr(1),
		},
		{
			name:     "pg parameter with equals",
			url:      "http://example.com/foo?pg=1",
			expected: intPtr(1),
		},
		{
			name:     "p parameter with equals",
			url:      "http://example.com/foo?p=1",
			expected: intPtr(1),
		},
		{
			name:     "paging parameter with equals",
			url:      "http://example.com/foo?paging=1",
			expected: intPtr(1),
		},
		{
			name:     "pag parameter with equals",
			url:      "http://example.com/foo?pag=1",
			expected: intPtr(1),
		},
		{
			name:     "pagination with slash separator",
			url:      "http://example.com/foo?pagination/1",
			expected: intPtr(1),
		},
		{
			name:     "paging with slash separator and large number",
			url:      "http://example.com/foo?paging/99",
			expected: intPtr(99),
		},
		{
			name:     "pa with slash separator",
			url:      "http://example.com/foo?pa/99",
			expected: intPtr(99),
		},
		{
			name:     "p with slash separator",
			url:      "http://example.com/foo?p/99",
			expected: intPtr(99),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PageNumFromURL(tt.url)
			
			if tt.expected == nil {
				if result != nil {
					t.Errorf("PageNumFromURL(%q) = %v, want nil", tt.url, *result)
				}
			} else {
				if result == nil {
					t.Errorf("PageNumFromURL(%q) = nil, want %d", tt.url, *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("PageNumFromURL(%q) = %d, want %d", tt.url, *result, *tt.expected)
				}
			}
		})
	}
}

// Helper function to create int pointers for test expectations
func intPtr(i int) *int {
	return &i
}