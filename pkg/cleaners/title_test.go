// ABOUTME: Comprehensive tests for title cleaner with JavaScript compatibility verification
// ABOUTME: Tests the CleanTitle function and supporting utilities for 100% JavaScript behavior

package cleaners

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestCleanTitle_BasicFunctionality tests the core title cleaning functionality
func TestCleanTitle_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		url      string
		expected string
	}{
		{
			name:     "simple title without separators",
			title:    "Simple Article Title",
			url:      "https://example.com/article",
			expected: "Simple Article Title",
		},
		{
			name:     "title with HTML tags",
			title:    "<strong>Important</strong> Article Title",
			url:      "https://example.com/article",
			expected: "Important Article Title",
		},
		{
			name:     "title with extra whitespace",
			title:    "  Article   Title   With   Spaces  ",
			url:      "https://example.com/article",
			expected: "Article Title With Spaces",
		},
		{
			name:     "title with site name using pipe separator",
			title:    "Article Title | Example.com",
			url:      "https://example.com/article",
			expected: "Article Title",
		},
		{
			name:     "title with site name using dash separator",
			title:    "Article Title - Example News",
			url:      "https://example.com/article",
			expected: "Article Title",
		},
		{
			name:     "title with site name using colon separator",
			title:    "Article Title: Example Blog",
			url:      "https://example.com/article",
			expected: "Article Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock document with an h1 for fallback testing
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
				<html>
					<head><title>` + tt.title + `</title></head>
					<body><h1>Fallback Title</h1></body>
				</html>
			`))
			if err != nil {
				t.Fatalf("Failed to create test document: %v", err)
			}

			result := CleanTitle(tt.title, tt.url, doc)
			if result != tt.expected {
				t.Errorf("CleanTitle(%q, %q) = %q, expected %q", tt.title, tt.url, result, tt.expected)
			}
		})
	}
}

// TestCleanTitle_LongTitleFallback tests the H1 fallback for overly long titles
func TestCleanTitle_LongTitleFallback(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		h1Text   string
		expected string
	}{
		{
			name:     "very long title falls back to h1",
			title:    "This is an extremely long title that exceeds the 150 character limit and should trigger the fallback mechanism to use the h1 element instead of the original title",
			h1Text:   "Short H1 Title",
			expected: "Short H1 Title",
		},
		{
			name:     "short title preserved",
			title:    "Short Title",
			h1Text:   "Alternative H1",
			expected: "Short Title",
		},
		{
			name:     "no h1 element keeps original long title",
			title:    "This is an extremely long title that exceeds the 150 character limit but there is no h1 element available so the original title should be kept",
			h1Text:   "",
			expected: "This is an extremely long title that exceeds the 150 character limit but there is no h1 element available so the original title should be kept",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			htmlContent := `<html><body>`
			if tt.h1Text != "" {
				htmlContent += `<h1>` + tt.h1Text + `</h1>`
			}
			htmlContent += `</body></html>`

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
			if err != nil {
				t.Fatalf("Failed to create test document: %v", err)
			}

			result := CleanTitle(tt.title, "https://example.com", doc)
			if result != tt.expected {
				t.Errorf("CleanTitle(%q) = %q, expected %q", tt.title, result, tt.expected)
			}
		})
	}
}

// TestResolveSplitTitle_BreadcrumbTitles tests complex breadcrumb title resolution
func TestResolveSplitTitle_BreadcrumbTitles(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		url      string
		expected string
	}{
		{
			name:     "simple breadcrumb with repeated separator",
			title:    "The Best Gadgets on Earth : Bits : Blogs : NYTimes.com",
			url:      "https://nytimes.com/article",
			expected: "The Best Gadgets on Earth",
		},
		{
			name:     "reverse breadcrumb pattern",
			title:    "NYTimes - Blogs - Bits - The Best Gadgets on Earth",
			url:      "https://nytimes.com/article",
			expected: "The Best Gadgets on Earth",
		},
		{
			name:     "title without domain match keeps original",
			title:    "Article Title - Site Name",
			url:      "https://example.com/article",
			expected: "Article Title - Site Name",
		},
		{
			name:     "title with similar domain name gets cleaned",
			title:    "Article Title - Examples News",
			url:      "https://example.com/article",
			expected: "Article Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveSplitTitle(tt.title, tt.url)
			if result != tt.expected {
				t.Errorf("ResolveSplitTitle(%q, %q) = %q, expected %q", tt.title, tt.url, result, tt.expected)
			}
		})
	}
}

// TestResolveSplitTitle_DomainCleaning tests fuzzy domain name removal
func TestResolveSplitTitle_DomainCleaning(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		url      string
		expected string
	}{
		{
			name:     "domain name at end of title",
			title:    "Great Article - Reddit",
			url:      "https://reddit.com/r/technology",
			expected: "Great Article",
		},
		{
			name:     "domain name at start of title",
			title:    "Reddit: Great Discussion Topic",
			url:      "https://reddit.com/r/technology",
			expected: "Great Discussion Topic",
		},
		{
			name:     "fuzzy domain match",
			title:    "Awesome Story | redditt",
			url:      "https://reddit.com/r/technology",
			expected: "Awesome Story",
		},
		{
			name:     "no domain match",
			title:    "Article Title - Different Site",
			url:      "https://example.com/article",
			expected: "Article Title - Different Site",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveSplitTitle(tt.title, tt.url)
			if result != tt.expected {
				t.Errorf("ResolveSplitTitle(%q, %q) = %q, expected %q", tt.title, tt.url, result, tt.expected)
			}
		})
	}
}

// TestExtractBreadcrumbTitle tests breadcrumb title extraction logic
func TestExtractBreadcrumbTitle(t *testing.T) {
	tests := []struct {
		name       string
		splitTitle []string
		text       string
		expected   string
	}{
		{
			name:       "breadcrumb with repeated separator",
			splitTitle: []string{"The Best Gadgets on Earth", " : ", "Bits", " : ", "Blogs", " : ", "NYTimes"},
			text:       "The Best Gadgets on Earth : Bits : Blogs : NYTimes",
			expected:   "The Best Gadgets on Earth",
		},
		{
			name:       "not enough segments",
			splitTitle: []string{"Title", " - ", "Site"},
			text:       "Title - Site",
			expected:   "",
		},
		{
			name:       "no repeated separators returns original text",
			splitTitle: []string{"One", " | ", "Two", " - ", "Three", " : ", "Four", " >> ", "Five"},
			text:       "One | Two - Three : Four >> Five",
			expected:   "One | Two - Three : Four >> Five",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractBreadcrumbTitle(tt.splitTitle, tt.text)
			if result != tt.expected {
				t.Errorf("ExtractBreadcrumbTitle() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestCleanDomainFromTitle tests domain removal with fuzzy matching
func TestCleanDomainFromTitle(t *testing.T) {
	tests := []struct {
		name       string
		splitTitle []string
		url        string
		expected   string
	}{
		{
			name:       "domain at start",
			splitTitle: []string{"reddit", " - ", "Great Article"},
			url:        "https://reddit.com/r/technology",
			expected:   "Great Article",
		},
		{
			name:       "domain at end",
			splitTitle: []string{"Great Article", " | ", "reddit"},
			url:        "https://reddit.com/r/technology",
			expected:   "Great Article",
		},
		{
			name:       "no domain match",
			splitTitle: []string{"Article", " - ", "Other"},
			url:        "https://example.com/article",
			expected:   "",
		},
		{
			name:       "invalid URL",
			splitTitle: []string{"Article", " - ", "Site"},
			url:        "invalid-url",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanDomainFromTitle(tt.splitTitle, tt.url)
			if result != tt.expected {
				t.Errorf("CleanDomainFromTitle() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestLevenshteinRatio tests the fuzzy string matching
func TestLevenshteinRatio(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected float64
	}{
		{
			name:     "identical strings",
			s1:       "reddit",
			s2:       "reddit",
			expected: 1.0,
		},
		{
			name:     "completely different strings",
			s1:       "reddit",
			s2:       "facebook",
			expected: 0.125, // (8-7)/8 = 0.125
		},
		{
			name:     "similar strings",
			s1:       "reddit",
			s2:       "redditt",
			expected: 0.857, // Approximately (7-1)/7
		},
		{
			name:     "empty strings",
			s1:       "",
			s2:       "test",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LevenshteinRatio(tt.s1, tt.s2)
			if abs(result-tt.expected) > 0.01 { // Allow small floating point differences
				t.Errorf("LevenshteinRatio(%q, %q) = %f, expected %f", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

// TestSplitTitleWithSeparators tests separator preservation during splitting
func TestSplitTitleWithSeparators(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected []string
	}{
		{
			name:     "title with pipe separator",
			title:    "Article Title | Site Name",
			expected: []string{"Article Title", " | ", "Site Name"},
		},
		{
			name:     "title with dash separator",
			title:    "Article Title - Site Name",
			expected: []string{"Article Title", " - ", "Site Name"},
		},
		{
			name:     "title with colon separator",
			title:    "Article Title: Site Name",
			expected: []string{"Article Title", ": ", "Site Name"},
		},
		{
			name:     "title with multiple separators",
			title:    "One | Two - Three: Four",
			expected: []string{"One", " | ", "Two", " - ", "Three", ": ", "Four"},
		},
		{
			name:     "title without separators",
			title:    "Simple Title",
			expected: []string{"Simple Title"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitTitleWithSeparators(tt.title)
			if len(result) != len(tt.expected) {
				t.Errorf("SplitTitleWithSeparators(%q) length = %d, expected %d", tt.title, len(result), len(tt.expected))
				return
			}
			for i, segment := range result {
				if segment != tt.expected[i] {
					t.Errorf("SplitTitleWithSeparators(%q)[%d] = %q, expected %q", tt.title, i, segment, tt.expected[i])
				}
			}
		})
	}
}

// TestCleanTitle_JavaScriptCompatibility tests for 100% JavaScript compatibility
func TestCleanTitle_JavaScriptCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		url      string
		expected string
		desc     string
	}{
		{
			name:     "real world reddit title",
			title:    "TIL something amazing happened | reddit",
			url:      "https://reddit.com/r/todayilearned/123",
			expected: "TIL something amazing happened",
			desc:     "Should remove fuzzy domain match at end",
		},
		{
			name:     "real world news title",
			title:    "Breaking News: Important Event - CNN.com",
			url:      "https://cnn.com/news/article",
			expected: "Breaking News: Important Event",
			desc:     "Should remove domain from end",
		},
		{
			name:     "complex breadcrumb title",
			title:    "Tech Review : Gadgets : Reviews : TechSite.com",
			url:      "https://techsite.com/reviews/gadgets",
			expected: "Tech Review",
			desc:     "Should extract main content from breadcrumb",
		},
		{
			name:     "title with HTML and separators",
			title:    "<strong>Amazing</strong> Article | <em>Example</em> News",
			url:      "https://example.com/news",
			expected: "Amazing Article",
			desc:     "Should clean HTML and remove site name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<html><body><h1>Fallback</h1></body></html>`))
			if err != nil {
				t.Fatalf("Failed to create test document: %v", err)
			}

			result := CleanTitle(tt.title, tt.url, doc)
			if result != tt.expected {
				t.Errorf("%s: CleanTitle(%q, %q) = %q, expected %q", tt.desc, tt.title, tt.url, result, tt.expected)
			}
		})
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}