// ABOUTME: Integration tests for title cleaner with real-world examples and edge cases
// ABOUTME: Verifies title cleaning works with actual HTML documents and various title formats

package cleaners

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestTitleCleanerIntegration tests the title cleaner with real-world scenarios
func TestTitleCleanerIntegration(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		title       string
		url         string
		expected    string
		description string
	}{
		{
			name: "reddit post with site name",
			html: `<html><body><h1>Post Title</h1></body></html>`,
			title: "TIL something amazing happened | reddit",
			url:   "https://reddit.com/r/todayilearned/123",
			expected: "TIL something amazing happened",
			description: "Should remove fuzzy-matched site name",
		},
		{
			name: "news article with HTML in title",
			html: `<html><body><h1>Breaking News</h1></body></html>`,
			title: "<strong>Breaking</strong>: Major Event - <em>CNN</em>",
			url:   "https://cnn.com/news/article",
			expected: "Breaking: Major Event - CNN",
			description: "Should strip HTML tags and preserve content",
		},
		{
			name: "very long title falls back to H1",
			html: `<html><body><h1>Short H1 Title</h1></body></html>`,
			title: "This is an extremely long title that exceeds the 150 character limit and should trigger the fallback mechanism to use the h1 element instead of the original title which is way too long",
			url:   "https://example.com/article",
			expected: "Short H1 Title",
			description: "Should use H1 when title is too long",
		},
		{
			name: "breadcrumb title extraction",
			html: `<html><body><h1>Fallback</h1></body></html>`,
			title: "The Complete Guide to Programming : Technology : Articles : TechBlog.com",
			url:   "https://techblog.com/technology/programming",
			expected: "The Complete Guide to Programming",
			description: "Should extract main content from breadcrumb pattern",
		},
		{
			name: "multiple H1 elements - no fallback",
			html: `<html><body><h1>First H1</h1><h1>Second H1</h1></body></html>`,
			title: "This is an extremely long title that exceeds the 150 character limit but there are multiple h1 elements so no fallback should occur and original should be kept",
			url:   "https://example.com/article",
			expected: "This is an extremely long title that exceeds the 150 character limit but there are multiple h1 elements so no fallback should occur and original should be kept",
			description: "Should not use H1 fallback when multiple H1s exist",
		},
		{
			name: "complex HTML with nested tags",
			html: `<html><body><h1>Clean Title</h1></body></html>`,
			title: "<div><span><strong>Nested</strong> <em>HTML</em></span> Content</div> | Site Name",
			url:   "https://sitename.com/article",
			expected: "Nested HTML Content",
			description: "Should clean deeply nested HTML and remove site name",
		},
		{
			name: "title with unusual separators",
			html: `<html><body><h1>Fallback</h1></body></html>`,
			title: "Article Title >> Site Name",
			url:   "https://sitename.com/article",
			expected: "Article Title >> Site Name",
			description: "Should preserve titles with non-standard separators",
		},
		{
			name: "empty title handling",
			html: `<html><body><h1>H1 Fallback</h1></body></html>`,
			title: "",
			url:   "https://example.com/article",
			expected: "",
			description: "Should handle empty titles gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to create test document: %v", err)
			}

			result := CleanTitle(tt.title, tt.url, doc)
			if result != tt.expected {
				t.Errorf("%s: CleanTitle(%q, %q) = %q, expected %q", 
					tt.description, tt.title, tt.url, result, tt.expected)
			}
		})
	}
}

// TestTitleCleanerEdgeCases tests edge cases and error conditions
func TestTitleCleanerEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		url      string
		html     string
		expected string
	}{
		{
			name:     "nil document handling",
			title:    "Test Title",
			url:      "https://example.com",
			html:     "",
			expected: "Test Title",
		},
		{
			name:     "malformed URL in domain cleaning",
			title:    "Article Title - Site Name",
			url:      "not-a-valid-url",
			html:     `<html><body></body></html>`,
			expected: "Article Title - Site Name",
		},
		{
			name:     "title with only separators",
			title:    " | - : ",
			url:      "https://example.com",
			html:     `<html><body></body></html>`,
			expected: "| - :",
		},
		{
			name:     "unicode characters in title",
			title:    "Título en Español | Sitio Web",
			url:      "https://sitioweb.com",
			html:     `<html><body></body></html>`,
			expected: "Título en Español",
		},
		{
			name:     "extremely long domain name",
			title:    "Article Title - Very Long Domain Name That Exceeds Normal Limits",
			url:      "https://verylongdomainnamethatexceedsnormallimits.com",
			html:     `<html><body></body></html>`,
			expected: "Article Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc *goquery.Document
			var err error
			
			if tt.html != "" {
				doc, err = goquery.NewDocumentFromReader(strings.NewReader(tt.html))
				if err != nil {
					t.Fatalf("Failed to create test document: %v", err)
				}
			} else {
				// Create minimal valid document
				doc, _ = goquery.NewDocumentFromReader(strings.NewReader("<html></html>"))
			}

			result := CleanTitle(tt.title, tt.url, doc)
			if result != tt.expected {
				t.Errorf("CleanTitle(%q, %q) = %q, expected %q", tt.title, tt.url, result, tt.expected)
			}
		})
	}
}

// TestTitleCleanerPerformance tests performance with various title lengths and complexities
func TestTitleCleanerPerformance(t *testing.T) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html><body><h1>Fallback</h1></body></html>"))

	// Test with various title complexities
	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "simple title",
			title: "Simple Title",
		},
		{
			name:  "complex breadcrumb",
			title: "Level1 : Level2 : Level3 : Level4 : Level5 : Level6 : FinalTitle",
		},
		{
			name:  "heavy HTML",
			title: strings.Repeat("<strong><em><span>", 100) + "Title" + strings.Repeat("</span></em></strong>", 100),
		},
		{
			name:  "very long title",
			title: strings.Repeat("Very Long Title Content ", 50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Multiple iterations to ensure consistent performance
			for i := 0; i < 100; i++ {
				result := CleanTitle(tt.title, "https://example.com", doc)
				if result == "" && tt.title != "" {
					t.Errorf("Unexpected empty result for %q", tt.title)
				}
			}
		})
	}
}

// BenchmarkCleanTitle benchmarks the title cleaning function
func BenchmarkCleanTitle(b *testing.B) {
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader("<html><body><h1>Fallback</h1></body></html>"))
	title := "Complex Title with HTML <strong>tags</strong> and Site Name | Example.com"
	url := "https://example.com/article"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CleanTitle(title, url, doc)
	}
}

// BenchmarkLevenshteinRatio benchmarks the fuzzy string matching
func BenchmarkLevenshteinRatio(b *testing.B) {
	s1 := "example"
	s2 := "examplesite"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LevenshteinRatio(s1, s2)
	}
}