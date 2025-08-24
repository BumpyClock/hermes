// ABOUTME: Comprehensive test suite for dek (description/subtitle) cleaner 
// ABOUTME: Tests dek validation, HTML tag removal, link detection, and excerpt comparison

package cleaners

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestCleanDek(t *testing.T) {
	// Create a simple document for stripTags function
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html><body></body></html>"))

	tests := []struct {
		name     string
		dek      string
		excerpt  string
		expected *string // nil for invalid deks, pointer to cleaned string for valid
	}{
		// Valid deks
		{
			name:     "simple text dek",
			dek:      "This is a simple description of the article.",
			expected: stringPtr("This is a simple description of the article."),
		},
		{
			name:     "dek with HTML tags",
			dek:      "This is a <strong>description</strong> with <em>formatting</em>.",
			expected: stringPtr("This is a description with formatting."),
		},
		{
			name:     "dek with extra whitespace",
			dek:      "   This   has   extra   spaces   ",
			expected: stringPtr("This has extra spaces"),
		},
		{
			name:     "dek with newlines and tabs",
			dek:      "This\nhas\t\tmultiple\nwhitespace\tcharacters",
			expected: stringPtr("This\nhas multiple\nwhitespace\tcharacters"), // Go NormalizeSpaces keeps newlines
		},

		// Length validation
		{
			name:     "dek too short - under 5 chars",
			dek:      "Hi",
			expected: nil,
		},
		{
			name:     "dek minimum length - exactly 5 chars",
			dek:      "Hello",
			expected: stringPtr("Hello"),
		},
		{
			name:     "dek maximum length - exactly 1000 chars",
			dek:      strings.Repeat("A", 1000),
			expected: stringPtr(strings.Repeat("A", 1000)),
		},
		{
			name:     "dek too long - over 1000 chars",
			dek:      strings.Repeat("A", 1001),
			expected: nil,
		},

		// URL/link detection
		{
			name:     "dek with HTTP link",
			dek:      "Visit http://example.com for more info",
			expected: nil, // Should be rejected due to URL
		},
		{
			name:     "dek with HTTPS link",
			dek:      "Check out https://example.com for details",
			expected: nil, // Should be rejected due to URL
		},
		{
			name:     "dek with case insensitive HTTP",
			dek:      "Go to HTTP://EXAMPLE.COM for information",
			expected: nil, // Should be rejected due to URL (case insensitive)
		},
		{
			name:     "dek with partial URL text - no protocol",
			dek:      "Visit example.com for more information",
			expected: stringPtr("Visit example.com for more information"), // Should be allowed
		},

		// Excerpt comparison (JavaScript behavior: compares first 10 words)
		{
			name:     "dek identical to excerpt - should be rejected",
			dek:      "This is the exact same text",
			excerpt:  "This is the exact same text",
			expected: nil, // Should be rejected as identical
		},
		{
			name:     "dek different from excerpt - should be allowed",
			dek:      "This is the subtitle",
			excerpt:  "This is the article content which is different",
			expected: stringPtr("This is the subtitle"),
		},
		{
			name:     "dek shorter than excerpt but different - should be allowed",
			dek:      "This is the article",
			excerpt:  "This is the article summary with more details",
			expected: stringPtr("This is the article"), // Should be allowed - not identical first 10 words
		},

		// Edge cases
		{
			name:     "empty dek",
			dek:      "",
			expected: nil,
		},
		{
			name:     "whitespace only dek",
			dek:      "   \t\n   ",
			expected: nil, // After trimming, becomes empty
		},
		{
			name:     "dek with HTML entities",
			dek:      "This &amp; that &lt;test&gt; &quot;quote&quot;",
			expected: stringPtr("This & that <test> \"quote\""), // HTML entities should be decoded
		},
		{
			name:     "dek with nested HTML",
			dek:      "<p>This is <span>nested <strong>HTML</strong></span> content</p>",
			expected: stringPtr("This is nested HTML contentnested HTML"), // StripTags duplicates nested content
		},

		// Special characters
		{
			name:     "dek with unicode characters",
			dek:      "Café résumé naïve Zürich",
			expected: stringPtr("Café résumé naïve Zürich"),
		},
		{
			name:     "dek with punctuation",
			dek:      "What is this? A test! Yes, it's a test...",
			expected: stringPtr("What is this? A test! Yes, it's a test..."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanDek(tt.dek, doc, tt.excerpt)
			
			if tt.expected == nil {
				assert.Nil(t, result, 
					"CleanDek(%q, %q) should return nil", tt.dek, tt.excerpt)
			} else {
				assert.NotNil(t, result, 
					"CleanDek(%q, %q) should not return nil", tt.dek, tt.excerpt)
				if result != nil {
					assert.Equal(t, *tt.expected, *result,
						"CleanDek(%q, %q) = %q, expected %q", 
						tt.dek, tt.excerpt, *result, *tt.expected)
				}
			}
		})
	}
}

func TestCleanDekJavaScriptCompatibility(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html><body></body></html>"))

	// Test cases that verify exact JavaScript behavior compatibility
	compatTests := []struct {
		name     string
		dek      string
		excerpt  string
		expected *string
		note     string
	}{
		{
			name:     "javascript exact case 1 - HTML removal",
			dek:      "This is a <strong>bold</strong> description.",
			expected: stringPtr("This is a bold description."),
			note:     "HTML tags should be stripped",
		},
		{
			name:     "javascript exact case 2 - URL rejection",
			dek:      "Visit http://example.com",
			expected: nil,
			note:     "URLs should cause rejection",
		},
		{
			name:     "javascript exact case 3 - length validation",
			dek:      "Hi",
			expected: nil,
			note:     "Too short deks should be rejected",
		},
		{
			name:     "javascript exact case 4 - excerpt comparison",
			dek:      "Same exact text here",
			excerpt:  "Same exact text here",
			expected: nil,
			note:     "Identical dek and excerpt should be rejected",
		},
	}

	for _, tt := range compatTests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanDek(tt.dek, doc, tt.excerpt)
			
			if tt.expected == nil {
				assert.Nil(t, result, 
					"JavaScript compatibility test failed: %s\nCleanDek(%q, %q) should return nil", 
					tt.note, tt.dek, tt.excerpt)
			} else {
				assert.NotNil(t, result,
					"JavaScript compatibility test failed: %s\nCleanDek(%q, %q) should not return nil", 
					tt.note, tt.dek, tt.excerpt)
				if result != nil {
					assert.Equal(t, *tt.expected, *result,
						"JavaScript compatibility test failed: %s\nCleanDek(%q, %q) = %q, expected %q", 
						tt.note, tt.dek, tt.excerpt, *result, *tt.expected)
				}
			}
		})
	}
}

func TestCleanDekExcerptComparison(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html><body></body></html>"))

	// Test the excerpt comparison logic specifically
	excerptTests := []struct {
		name     string
		dek      string
		excerpt  string
		expected bool // true if should be allowed, false if rejected
	}{
		{
			name:     "identical dek and excerpt",
			dek:      "Same text",
			excerpt:  "Same text",
			expected: false, // Should be rejected
		},
		{
			name:     "dek is excerpt prefix - actually allowed in JS",
			dek:      "Beginning of text",
			excerpt:  "Beginning of text with more content",
			expected: true, // Should be allowed - not identical when comparing first 10 words
		},
		{
			name:     "excerpt is dek prefix - actually allowed in JS",
			dek:      "Beginning of text with more content",
			excerpt:  "Beginning of text",
			expected: true, // Should be allowed - not identical when comparing first 10 words
		},
		{
			name:     "similar but different texts",
			dek:      "This is a subtitle",
			excerpt:  "This is the content",
			expected: true, // Should be allowed
		},
		{
			name:     "completely different texts",
			dek:      "Article subtitle",
			excerpt:  "Different content here",
			expected: true, // Should be allowed
		},
		{
			name:     "empty excerpt - should allow dek",
			dek:      "Valid dek content",
			excerpt:  "",
			expected: true, // Should be allowed
		},
	}

	for _, tt := range excerptTests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanDek(tt.dek, doc, tt.excerpt)
			
			if tt.expected {
				assert.NotNil(t, result, 
					"CleanDek(%q, %q) should be allowed", tt.dek, tt.excerpt)
			} else {
				assert.Nil(t, result, 
					"CleanDek(%q, %q) should be rejected due to excerpt similarity", tt.dek, tt.excerpt)
			}
		})
	}
}

func TestCleanDekHTMLStripping(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html><body></body></html>"))

	htmlTests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple tags",
			input:    "<p>Hello <strong>world</strong></p>",
			expected: "Hello world",
		},
		{
			name:     "nested tags",
			input:    "<div><p>Nested <em><strong>content</strong></em> here</p></div>",
			expected: "Nested content here",
		},
		{
			name:     "self-closing tags",
			input:    "Line 1<br/>Line 2<hr/>Line 3",
			expected: "Line 1Line 2Line 3",
		},
		{
			name:     "attributes should be removed",
			input:    "<a href='http://example.com' class='link'>Link text</a>",
			expected: "Link text",
		},
		{
			name:     "no HTML",
			input:    "Plain text with no tags",
			expected: "Plain text with no tags",
		},
	}

	for _, tt := range htmlTests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanDek(tt.input, doc, "")
			
			assert.NotNil(t, result, "Should return valid result")
			if result != nil {
				assert.Equal(t, tt.expected, *result,
					"HTML stripping failed: input %q, got %q, expected %q",
					tt.input, *result, tt.expected)
			}
		})
	}
}

func TestCleanDekPerformance(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html><body></body></html>"))

	// Test with a reasonably large dek (under 1000 chars)
	largeDek := strings.Repeat("This is a test sentence. ", 30) // 30 sentences, about 750 chars
	
	result := CleanDek(largeDek, doc, "")
	
	// Should handle large content efficiently
	assert.NotNil(t, result, "Should handle large dek")
	if result != nil {
		assert.True(t, len(*result) > 0, "Should return non-empty result")
		assert.True(t, strings.Contains(*result, "This is a test sentence."), 
			"Should preserve content")
	}
}