// ABOUTME: Comprehensive test suite for URL extractor with JavaScript compatibility verification
// ABOUTME: Tests canonical link detection, OpenGraph URL extraction, and domain parsing

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestGenericUrlExtractor_Basic tests the basic functionality matching JavaScript tests
func TestGenericUrlExtractor_Basic(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		url         string
		metaCache   []string
		expectedURL string
		expectedDomain string
	}{
		{
			name: "returns canonical url and domain first",
			html: `
				<html>
					<head>
						<link rel="canonical" href="https://example.com/blog/post" />
						<meta name="og:url" value="https://example.com/blog/post" />
					</head>
				</html>
			`,
			url:            "https://example.com/blog/post?utm_campain=poajwefpaoiwjefaepoj",
			metaCache:      []string{"og:url"},
			expectedURL:    "https://example.com/blog/post",
			expectedDomain: "example.com",
		},
		{
			name: "returns og:url second",
			html: `
				<html>
					<head>
						<meta name="og:url" value="https://example.com/blog/post" />
					</head>
				</html>
			`,
			url:            "https://example.com/blog/post?utm_campain=poajwefpaoiwjefaepoj",
			metaCache:      []string{"og:url"},
			expectedURL:    "https://example.com/blog/post",
			expectedDomain: "example.com",
		},
		{
			name: "returns passed url if others are not found",
			html: `
				<html>
					<head>
					</head>
				</html>
			`,
			url:            "https://example.com/blog/post?utm_campain=poajwefpaoiwjefaepoj",
			metaCache:      []string{},
			expectedURL:    "https://example.com/blog/post?utm_campain=poajwefpaoiwjefaepoj",
			expectedDomain: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			result := GenericUrlExtractor.Extract(doc.Selection, tt.url, tt.metaCache)

			if result.URL != tt.expectedURL {
				t.Errorf("Expected URL %q, got %q", tt.expectedURL, result.URL)
			}

			if result.Domain != tt.expectedDomain {
				t.Errorf("Expected domain %q, got %q", tt.expectedDomain, result.Domain)
			}
		})
	}
}

// TestParseDomain tests the domain parsing function
func TestParseDomain(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "basic domain",
			url:      "https://example.com/path",
			expected: "example.com",
		},
		{
			name:     "subdomain",
			url:      "https://www.example.com/path",
			expected: "www.example.com",
		},
		{
			name:     "port number",
			url:      "https://example.com:8080/path",
			expected: "example.com",
		},
		{
			name:     "complex subdomain",
			url:      "https://api.v2.example.co.uk/path",
			expected: "api.v2.example.co.uk",
		},
		{
			name:     "IP address",
			url:      "https://192.168.1.1:3000/path",
			expected: "192.168.1.1",
		},
		{
			name:     "localhost",
			url:      "http://localhost:8080/path",
			expected: "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDomain(tt.url)
			if result != tt.expected {
				t.Errorf("Expected domain %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestGenericUrlExtractor_CanonicalPriority tests canonical link priority over meta tags
func TestGenericUrlExtractor_CanonicalPriority(t *testing.T) {
	html := `
		<html>
			<head>
				<link rel="canonical" href="https://example.com/canonical" />
				<meta name="og:url" value="https://example.com/og-url" />
			</head>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := GenericUrlExtractor.Extract(doc.Selection, "https://example.com/original", []string{"og:url"})

	// Canonical link should win over meta tag
	if result.URL != "https://example.com/canonical" {
		t.Errorf("Expected canonical URL, got %q", result.URL)
	}
}

// TestGenericUrlExtractor_EmptyCanonical tests handling of empty canonical href
func TestGenericUrlExtractor_EmptyCanonical(t *testing.T) {
	html := `
		<html>
			<head>
				<link rel="canonical" href="" />
				<meta name="og:url" value="https://example.com/og-url" />
			</head>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := GenericUrlExtractor.Extract(doc.Selection, "https://example.com/original", []string{"og:url"})

	// Should fallback to og:url when canonical is empty
	if result.URL != "https://example.com/og-url" {
		t.Errorf("Expected og:url fallback, got %q", result.URL)
	}
}

// TestGenericUrlExtractor_MultipleCanonical tests handling of multiple canonical links
func TestGenericUrlExtractor_MultipleCanonical(t *testing.T) {
	html := `
		<html>
			<head>
				<link rel="canonical" href="https://example.com/first" />
				<link rel="canonical" href="https://example.com/second" />
			</head>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := GenericUrlExtractor.Extract(doc.Selection, "https://example.com/original", []string{})

	// Should use first canonical link
	if result.URL != "https://example.com/first" {
		t.Errorf("Expected first canonical URL, got %q", result.URL)
	}
}

// TestGenericUrlExtractor_NoCanonicalInCache tests meta tag extraction when not in cache
func TestGenericUrlExtractor_NoCanonicalInCache(t *testing.T) {
	html := `
		<html>
			<head>
				<meta name="og:url" value="https://example.com/og-url" />
			</head>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// og:url not in cache - should fallback to original URL
	result := GenericUrlExtractor.Extract(doc.Selection, "https://example.com/original", []string{})

	if result.URL != "https://example.com/original" {
		t.Errorf("Expected original URL when meta not in cache, got %q", result.URL)
	}
}

// TestGenericUrlExtractor_RelativeCanonical tests handling of relative canonical URLs
func TestGenericUrlExtractor_RelativeCanonical(t *testing.T) {
	html := `
		<html>
			<head>
				<link rel="canonical" href="/blog/post" />
			</head>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := GenericUrlExtractor.Extract(doc.Selection, "https://example.com/original", []string{})

	// Should return relative URL as-is (matching JavaScript behavior)
	if result.URL != "/blog/post" {
		t.Errorf("Expected relative canonical URL, got %q", result.URL)
	}

	// Domain should be extracted from relative URL (will be empty)
	if result.Domain != "" {
		t.Errorf("Expected empty domain for relative URL, got %q", result.Domain)
	}
}

// TestGenericUrlExtractor_JavaScriptCompatibility verifies exact JavaScript matching
func TestGenericUrlExtractor_JavaScriptCompatibility(t *testing.T) {
	// Test case matching the exact JavaScript test
	fullUrl := "https://example.com/blog/post?utm_campain=poajwefpaoiwjefaepoj"
	clean := "https://example.com/blog/post"

	html := `
		<html>
			<head>
				<link rel="canonical" href="` + clean + `" />
				<meta name="og:url" value="` + clean + `" />
			</head>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := GenericUrlExtractor.Extract(doc.Selection, fullUrl, []string{"og:url"})

	// Should match JavaScript exactly
	if result.URL != clean {
		t.Errorf("JavaScript compatibility: expected URL %q, got %q", clean, result.URL)
	}

	if result.Domain != "example.com" {
		t.Errorf("JavaScript compatibility: expected domain %q, got %q", "example.com", result.Domain)
	}
}

// TestParseDomain_EdgeCases tests edge cases for domain parsing
func TestParseDomain_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "empty URL",
			url:      "",
			expected: "",
		},
		{
			name:     "invalid URL",
			url:      "not-a-url",
			expected: "",
		},
		{
			name:     "no protocol",
			url:      "//example.com/path",
			expected: "example.com",
		},
		{
			name:     "protocol relative",
			url:      "//www.example.com/path",
			expected: "www.example.com",
		},
		{
			name:     "file protocol",
			url:      "file:///path/to/file.html",
			expected: "",
		},
		{
			name:     "ftp protocol",
			url:      "ftp://files.example.com/file.txt",
			expected: "files.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDomain(tt.url)
			if result != tt.expected {
				t.Errorf("Expected domain %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestGenericUrlExtractor_ComplexHTML tests extraction from complex HTML documents
func TestGenericUrlExtractor_ComplexHTML(t *testing.T) {
	html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>Test Article</title>
			<link rel="stylesheet" href="styles.css">
			<link rel="canonical" href="https://news.example.com/articles/breaking-news" />
			<meta property="og:url" content="https://news.example.com/articles/breaking-news" />
			<meta name="twitter:url" content="https://twitter.example.com/articles/breaking-news" />
		</head>
		<body>
			<article>
				<h1>Breaking News</h1>
				<p>This is the article content.</p>
			</article>
		</body>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := GenericUrlExtractor.Extract(doc.Selection, "https://news.example.com/articles/breaking-news?ref=homepage", []string{"og:url", "twitter:url"})

	expectedURL := "https://news.example.com/articles/breaking-news"
	expectedDomain := "news.example.com"

	if result.URL != expectedURL {
		t.Errorf("Expected URL %q, got %q", expectedURL, result.URL)
	}

	if result.Domain != expectedDomain {
		t.Errorf("Expected domain %q, got %q", expectedDomain, result.Domain)
	}
}

// BenchmarkGenericUrlExtractor benchmarks the URL extraction performance
func BenchmarkGenericUrlExtractor(b *testing.B) {
	html := `
		<html>
			<head>
				<link rel="canonical" href="https://example.com/blog/post" />
				<meta name="og:url" value="https://example.com/blog/post" />
			</head>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		b.Fatalf("Failed to parse HTML: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenericUrlExtractor.Extract(doc.Selection, "https://example.com/original", []string{"og:url"})
	}
}

// BenchmarkParseDomain benchmarks the domain parsing performance
func BenchmarkParseDomain(b *testing.B) {
	testURL := "https://www.example.com:8080/complex/path/to/resource?param=value#fragment"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseDomain(testURL)
	}
}