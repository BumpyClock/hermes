// ABOUTME: Comprehensive test suite for author cleaner with JavaScript compatibility verification
// ABOUTME: Tests all author cleaning scenarios including byline removal and normalization

package cleaners

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanAuthor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic "By" prefix removal
		{
			name:     "simple by prefix",
			input:    "By David Smith",
			expected: "David Smith",
		},
		{
			name:     "by prefix lowercase",
			input:    "by jane doe",
			expected: "jane doe",
		},
		{
			name:     "By with colon",
			input:    "By: John Johnson",
			expected: "John Johnson",
		},
		{
			name:     "by with colon and spaces",
			input:    "by :  Mary Jane",
			expected: "Mary Jane",
		},
		
		// "Posted by" and "Written by" variants
		{
			name:     "posted by prefix",
			input:    "Posted by Admin User",
			expected: "Admin User",
		},
		{
			name:     "written by prefix",
			input:    "Written by: Sarah Wilson",
			expected: "Sarah Wilson",
		},
		{
			name:     "posted by with extra spaces",
			input:    "  Posted by   Alice Cooper   ",
			expected: "Alice Cooper",
		},
		
		// Whitespace handling
		{
			name:     "leading and trailing spaces",
			input:    "   By David Smith   ",
			expected: "David Smith",
		},
		{
			name:     "multiple internal spaces",
			input:    "By  David    Smith",
			expected: "David Smith",
		},
		{
			name:     "tabs and newlines - JS behavior",
			input:    "By\tDavid\nSmith",
			expected: "David", // JavaScript .* stops at newlines
		},
		
		// No prefix cases
		{
			name:     "author without prefix",
			input:    "David Smith",
			expected: "David Smith",
		},
		{
			name:     "author with whitespace only",
			input:    "  David Smith  ",
			expected: "David Smith",
		},
		
		// Edge cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "by only",
			input:    "By",
			expected: "",
		},
		{
			name:     "by with colon only",
			input:    "By:",
			expected: "",
		},
		
		// Multiple authors
		{
			name:     "multiple authors",
			input:    "By John Smith and Jane Doe",
			expected: "John Smith and Jane Doe",
		},
		{
			name:     "comma separated authors",
			input:    "By John Smith, Jane Doe, Bob Wilson",
			expected: "John Smith, Jane Doe, Bob Wilson",
		},
		
		// Special characters
		{
			name:     "author with special chars",
			input:    "By Jean-Luc Picard",
			expected: "Jean-Luc Picard",
		},
		{
			name:     "author with apostrophe",
			input:    "By O'Reilly",
			expected: "O'Reilly",
		},
		{
			name:     "author with unicode",
			input:    "By José María",
			expected: "José María",
		},
		
		// Case sensitivity tests
		{
			name:     "BY uppercase",
			input:    "BY DAVID SMITH",
			expected: "DAVID SMITH",
		},
		{
			name:     "Mixed case prefix",
			input:    "PoStEd By Mixed Case",
			expected: "Mixed Case",
		},
		
		// Complex whitespace patterns
		{
			name:     "multiple consecutive spaces",
			input:    "By     David     Smith     ",
			expected: "David Smith",
		},
		{
			name:     "mixed whitespace types - JS behavior",
			input:    "By\n\tDavid  \r Smith\t\n",
			expected: "David", // JavaScript captures only "David  " before \r
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanAuthor(tt.input)
			assert.Equal(t, tt.expected, result, 
				"CleanAuthor(%q) = %q, expected %q", tt.input, result, tt.expected)
		})
	}
}

func TestCleanAuthorJavaScriptCompatibility(t *testing.T) {
	// Test cases that verify exact JavaScript behavior compatibility
	compatTests := []struct {
		name     string
		input    string
		expected string
		note     string
	}{
		{
			name:     "javascript exact case 1",
			input:    "By David Smith ",
			expected: "David Smith",
			note:     "Trailing space should be trimmed",
		},
		{
			name:     "javascript exact case 2", 
			input:    "posted by: John Doe",
			expected: "John Doe",
			note:     "Posted by with colon",
		},
		{
			name:     "javascript exact case 3",
			input:    "WRITTEN BY ADMIN",
			expected: "ADMIN",
			note:     "Case insensitive matching but preserve result case",
		},
		{
			name:     "javascript exact case 4",
			input:    "  by  :  Multiple   Spaces  ",
			expected: "Multiple Spaces",
			note:     "Complex whitespace normalization",
		},
	}

	for _, tt := range compatTests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanAuthor(tt.input)
			assert.Equal(t, tt.expected, result, 
				"JavaScript compatibility test failed: %s\nCleanAuthor(%q) = %q, expected %q", 
				tt.note, tt.input, result, tt.expected)
		})
	}
}

func TestCleanAuthorRegexPattern(t *testing.T) {
	// Test the regex pattern directly to ensure it matches JavaScript behavior
	regexTests := []struct {
		input    string
		expected []string // [full_match, prefix_group, author_group] 
	}{
		{
			input:    "By David Smith",
			expected: []string{"By David Smith", "", "David Smith"},
		},
		{
			input:    "Posted by: John Doe",
			expected: []string{"Posted by: John Doe", "Posted ", "John Doe"},
		},
		{
			input:    "written by Author Name",
			expected: []string{"written by Author Name", "written ", "Author Name"},
		},
		{
			input:    "by: Someone",
			expected: []string{"by: Someone", "", "Someone"},
		},
	}

	for _, tt := range regexTests {
		t.Run("regex_"+strings.ReplaceAll(tt.input, " ", "_"), func(t *testing.T) {
			matches := CLEAN_AUTHOR_RE.FindStringSubmatch(tt.input)
			
			if len(tt.expected) == 0 {
				assert.Nil(t, matches, "Expected no match for %q", tt.input)
			} else {
				assert.NotNil(t, matches, "Expected match for %q", tt.input)
				assert.Equal(t, len(tt.expected), len(matches), 
					"Wrong number of capture groups for %q", tt.input)
				
				for i, expected := range tt.expected {
					assert.Equal(t, expected, matches[i], 
						"Capture group %d mismatch for %q", i, tt.input)
				}
			}
		})
	}
}

func TestCleanAuthorPerformance(t *testing.T) {
	// Test performance with longer author strings
	longAuthor := "By " + strings.Repeat("Very Long Author Name ", 100)
	
	result := CleanAuthor(longAuthor)
	
	// Should still work correctly with long strings
	expected := strings.Repeat("Very Long Author Name ", 100)
	expected = strings.TrimSpace(expected)
	assert.Equal(t, expected, result)
}

func TestCleanAuthorEdgeCases(t *testing.T) {
	edgeCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "by at end of string",
			input:    "Author Name by",
			expected: "Author Name by", // Should not match
		},
		{
			name:     "by in middle of string", 
			input:    "Written by John by Smith",
			expected: "John by Smith", // Only matches prefix
		},
		{
			name:     "multiple by prefixes",
			input:    "by by John Smith",
			expected: "by John Smith", // Only first match
		},
		{
			name:     "by without space - JS actually matches",
			input:    "ByJohn Smith",
			expected: "John Smith", // JavaScript DOES match this
		},
		{
			name:     "newline in author - JS behavior",
			input:    "By John\nSmith",
			expected: "John", // JavaScript .* stops at newline
		},
	}

	for _, tt := range edgeCases {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanAuthor(tt.input)
			assert.Equal(t, tt.expected, result,
				"Edge case %s failed: CleanAuthor(%q) = %q, expected %q",
				tt.name, tt.input, result, tt.expected)
		})
	}
}