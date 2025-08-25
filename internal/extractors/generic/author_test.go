// ABOUTME: Port of extractors/generic/author/extractor.js test suite to Go
// This file provides 100% JavaScript-compatible author extraction testing
// with comprehensive coverage of meta tags, selectors, and byline patterns.
//
// JavaScript Compatibility: Tests verify exact behavior matching including:
// - Three-tier extraction strategy (meta -> selectors -> byline regex)
// - Author length limits and cleaning patterns
// - CSS selector prioritization and fallback logic
// - Byline regex pattern matching with case-insensitive 'By' detection
//
// Test Coverage: Meta tag extraction, CSS selectors, regex patterns, edge cases
// Performance: Optimized Go implementation with benchmark testing

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGenericAuthorExtractor_ExtractFromMeta(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "Extract from byl meta tag",
			html: `<html><head><meta name="byl" value="John Smith"></head><body></body></html>`,
			expected: "John Smith",
		},
		{
			name: "Extract from dc.author meta tag",
			html: `<html><head><meta name="dc.author" value="Jane Doe"></head><body></body></html>`,
			expected: "Jane Doe",
		},
		{
			name: "Extract from authors meta tag",
			html: `<html><head><meta name="authors" value="Bob Wilson"></head><body></body></html>`,
			expected: "Bob Wilson",
		},
		{
			name: "Priority order - byl over dc.author",
			html: `<html><head>
				<meta name="dc.author" value="Second Author">
				<meta name="byl" value="First Author">
			</head><body></body></html>`,
			expected: "First Author",
		},
		{
			name: "Author too long - should fall through",
			html: `<html><head><meta name="byl" value="` + strings.Repeat("Very Long Author Name ", 20) + `"></head><body></body></html>`,
			expected: "",
		},
		{
			name: "No meta tags",
			html: `<html><head></head><body></body></html>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			metaCache := []string{"byl", "dc.author", "authors", "clmst", "dcsext.author", "dc.creator", "rbauthors"}
			extractor := &GenericAuthorExtractor{}
			result := extractor.Extract(doc.Selection, metaCache)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected nil result, got %q", *result)
				}
			} else {
				if result == nil {
					t.Fatalf("Expected %q, got nil", tt.expected)
				}
				if *result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, *result)
				}
			}
		})
	}
}

func TestGenericAuthorExtractor_ExtractFromSelectors(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "Extract from .author class",
			html: `<html><body><div class="author">John Smith</div></body></html>`,
			expected: "John Smith",
		},
		{
			name: "Extract from .byline class",
			html: `<html><body><div class="byline">Jane Doe</div></body></html>`,
			expected: "Jane Doe",
		},
		{
			name: "Extract from vcard pattern",
			html: `<html><body><div class="author vcard"><span class="fn">Bob Wilson</span></div></body></html>`,
			expected: "Bob Wilson",
		},
		{
			name: "Extract from rel=author link",
			html: `<html><body><a rel="author" href="/author/john">John Smith</a></body></html>`,
			expected: "John Smith",
		},
		{
			name: "Priority order - .entry .entry-author over .author",
			html: `<html><body>
				<div class="author">Second Author</div>
				<div class="entry"><div class="entry-author">First Author</div></div>
			</body></html>`,
			expected: "First Author",
		},
		{
			name: "Author too long - should fall through",
			html: `<html><body><div class="author">` + strings.Repeat("Very Long Author Name ", 20) + `</div></body></html>`,
			expected: "",
		},
		{
			name: "No matching selectors",
			html: `<html><body><div class="content">Some content</div></body></html>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			metaCache := []string{"byl", "dc.author", "authors", "clmst", "dcsext.author", "dc.creator", "rbauthors"}
			extractor := &GenericAuthorExtractor{}
			result := extractor.Extract(doc.Selection, metaCache)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected nil result, got %q", *result)
				}
			} else {
				if result == nil {
					t.Fatalf("Expected %q, got nil", tt.expected)
				}
				if *result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, *result)
				}
			}
		})
	}
}

func TestGenericAuthorExtractor_ExtractFromBylineRegex(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "Extract from #byline with By pattern",
			html: `<html><body><div id="byline">By John Smith</div></body></html>`,
			expected: "John Smith",
		},
		{
			name: "Extract from .byline with By pattern",
			html: `<html><body><div class="byline">By Jane Doe</div></body></html>`,
			expected: "Jane Doe",
		},
		{
			name: "Case insensitive By pattern",
			html: `<html><body><div id="byline">BY BOB WILSON</div></body></html>`,
			expected: "BOB WILSON",
		},
		{
			name: "By pattern with whitespace and newlines",
			html: `<html><body><div id="byline">
			
			By   Sarah Johnson
			
			</div></body></html>`,
			expected: "Sarah Johnson",
		},
		{
			name: "Multiple byline elements - CSS selector takes priority over regex",
			html: `<html><body>
				<div id="byline">By First Author</div>
				<div class="byline">By Second Author</div>
			</body></html>`,
			expected: "Second Author",
		},
		{
			name: "Regex priority - #byline over .byline when CSS selectors disabled",
			html: `<html><body>
				<div id="byline">By First Author</div>
				<div class="different-class">By Second Author</div>
			</body></html>`,
			expected: "First Author",
		},
		{
			name: "Byline without By pattern - should not match",
			html: `<html><body><div id="byline">Just an author name</div></body></html>`,
			expected: "",
		},
		{
			name: "No byline elements",
			html: `<html><body><div class="content">Some content</div></body></html>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			metaCache := []string{"byl", "dc.author", "authors", "clmst", "dcsext.author", "dc.creator", "rbauthors"}
			extractor := &GenericAuthorExtractor{}
			result := extractor.Extract(doc.Selection, metaCache)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected nil result, got %q", *result)
				}
			} else {
				if result == nil {
					t.Fatalf("Expected %q, got nil", tt.expected)
				}
				if *result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, *result)
				}
			}
		})
	}
}

func TestGenericAuthorExtractor_ExtractionPriority(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "Meta tag takes priority over selectors",
			html: `<html><head><meta name="byl" value="Meta Author"></head>
				<body><div class="author">Selector Author</div></body></html>`,
			expected: "Meta Author",
		},
		{
			name: "Selectors take priority over byline regex",
			html: `<html><body>
				<div class="author">Selector Author</div>
				<div id="byline">By Regex Author</div>
			</body></html>`,
			expected: "Selector Author",
		},
		{
			name: "Falls through all strategies when no match",
			html: `<html><body><div class="content">No author info</div></body></html>`,
			expected: "",
		},
		{
			name: "Falls through when meta author too long",
			html: `<html><head><meta name="byl" value="` + strings.Repeat("Very Long Name ", 20) + `"></head>
				<body><div class="author">Good Author</div></body></html>`,
			expected: "Good Author",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			metaCache := []string{"byl", "dc.author", "authors", "clmst", "dcsext.author", "dc.creator", "rbauthors"}
			extractor := &GenericAuthorExtractor{}
			result := extractor.Extract(doc.Selection, metaCache)

			if tt.expected == "" {
				if result != nil {
					t.Errorf("Expected nil result, got %q", *result)
				}
			} else {
				if result == nil {
					t.Fatalf("Expected %q, got nil", tt.expected)
				}
				if *result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, *result)
				}
			}
		})
	}
}

func TestCleanAuthor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Clean 'By' prefix",
			input:    "By John Smith",
			expected: "John Smith",
		},
		{
			name:     "Clean 'by' prefix case insensitive",
			input:    "by jane doe",
			expected: "jane doe",
		},
		{
			name:     "Clean 'BY' prefix uppercase",
			input:    "BY BOB WILSON",
			expected: "BOB WILSON",
		},
		{
			name:     "Clean 'posted by' prefix",
			input:    "posted by Sarah Johnson",
			expected: "Sarah Johnson",
		},
		{
			name:     "Clean 'written by' prefix",
			input:    "written by Mike Davis",
			expected: "Mike Davis",
		},
		{
			name:     "Clean with colon",
			input:    "By: Emily Chen",
			expected: "Emily Chen",
		},
		{
			name:     "Clean with extra whitespace",
			input:    "  By   David   Miller  ",
			expected: "David Miller",
		},
		{
			name:     "Clean with newlines and tabs",
			input:    "\n\t By\t\nLisa   Brown  \n",
			expected: "Lisa Brown",
		},
		{
			name:     "No cleaning needed",
			input:    "John Smith",
			expected: "John Smith",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only 'By' prefix",
			input:    "By",
			expected: "",
		},
		{
			name:     "Complex cleaning",
			input:    "  posted by: Author Name  ",
			expected: "Author Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanAuthor(tt.input)
			if result != tt.expected {
				t.Errorf("cleanAuthor(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenericAuthorExtractor_Integration(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "Real-world example with meta and byline",
			html: `<html>
				<head><meta name="dc.author" value="By News Reporter"></head>
				<body>
					<article>
						<div class="byline">By Staff Writer</div>
						<div class="content">Article content here</div>
					</article>
				</body>
			</html>`,
			expected: "News Reporter",
		},
		{
			name: "Complex author extraction with cleaning",
			html: `<html>
				<body>
					<div class="entry">
						<div class="entry-author">  posted by: John Smith  </div>
					</div>
				</body>
			</html>`,
			expected: "John Smith",
		},
		{
			name: "Vcard microformat extraction",
			html: `<html>
				<body>
					<div class="author vcard">
						<span class="fn">Jane Doe</span>
					</div>
				</body>
			</html>`,
			expected: "Jane Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			metaCache := []string{"byl", "dc.author", "authors", "clmst", "dcsext.author", "dc.creator", "rbauthors"}
			extractor := &GenericAuthorExtractor{}
			result := extractor.Extract(doc.Selection, metaCache)

			if result == nil {
				t.Fatalf("Expected %q, got nil", tt.expected)
			}
			if *result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, *result)
			}
		})
	}
}

func BenchmarkGenericAuthorExtractor_Extract(b *testing.B) {
	html := `<html>
		<head><meta name="dc.author" value="Benchmark Author"></head>
		<body>
			<div class="author">Another Author</div>
			<div id="byline">By Third Author</div>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		b.Fatalf("Failed to parse HTML: %v", err)
	}

	metaCache := []string{"byl", "dc.author", "authors", "clmst", "dcsext.author", "dc.creator", "rbauthors"}
	extractor := &GenericAuthorExtractor{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.Extract(doc.Selection, metaCache)
	}
}

func BenchmarkCleanAuthor(b *testing.B) {
	input := "  posted by: John Smith  "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cleanAuthor(input)
	}
}