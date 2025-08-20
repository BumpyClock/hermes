// ABOUTME: Integration tests for URL extractor to verify it works with the existing parser system
// ABOUTME: Tests realistic scenarios with full HTML documents and various URL configurations

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestURLExtractorIntegration tests integration with realistic HTML
func TestURLExtractorIntegration(t *testing.T) {
	// Realistic HTML from a news article
	html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>Breaking: Major Tech Announcement | TechNews</title>
			<meta name="description" content="Latest tech news">
			<meta name="og:url" content="https://technews.example.com/breaking-tech-announcement">
			<meta name="og:title" content="Breaking: Major Tech Announcement">
			<meta name="twitter:url" content="https://twitter.technews.example.com/breaking">
			<link rel="canonical" href="https://technews.example.com/2024/01/breaking-tech-announcement">
			<link rel="stylesheet" href="/css/main.css">
		</head>
		<body>
			<header>
				<nav><!-- navigation --></nav>
			</header>
			<main>
				<article>
					<h1>Breaking: Major Tech Announcement</h1>
					<p class="byline">By John Smith, Tech Reporter</p>
					<time datetime="2024-01-19">January 19, 2024</time>
					<p>This is the article content...</p>
				</article>
			</main>
		</body>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Simulate meta cache that would be built by parser
	metaCache := []string{"og:url", "og:title", "twitter:url"}

	// Original URL with tracking parameters
	originalURL := "https://technews.example.com/2024/01/breaking-tech-announcement?utm_source=homepage&ref=social"

	result := GenericUrlExtractor.Extract(doc.Selection, originalURL, metaCache)

	// Should prioritize canonical link over meta tags
	expectedURL := "https://technews.example.com/2024/01/breaking-tech-announcement"
	expectedDomain := "technews.example.com"

	if result.URL != expectedURL {
		t.Errorf("Expected canonical URL %q, got %q", expectedURL, result.URL)
	}

	if result.Domain != expectedDomain {
		t.Errorf("Expected domain %q, got %q", expectedDomain, result.Domain)
	}
}

// TestURLExtractorWithoutCanonical tests fallback to meta tags
func TestURLExtractorWithoutCanonical(t *testing.T) {
	html := `
		<html>
		<head>
			<meta name="og:url" content="https://blog.example.com/clean-url">
			<meta name="twitter:url" content="https://blog.example.com/twitter-url">
		</head>
		<body><p>Content</p></body>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	metaCache := []string{"og:url", "twitter:url"}
	originalURL := "https://blog.example.com/ugly-url?session=12345&track=true"

	result := GenericUrlExtractor.Extract(doc.Selection, originalURL, metaCache)

	// Should use og:url since it's first in CANONICAL_META_SELECTORS
	expectedURL := "https://blog.example.com/clean-url"
	expectedDomain := "blog.example.com"

	if result.URL != expectedURL {
		t.Errorf("Expected og:url fallback %q, got %q", expectedURL, result.URL)
	}

	if result.Domain != expectedDomain {
		t.Errorf("Expected domain %q, got %q", expectedDomain, result.Domain)
	}
}

// TestURLExtractorEmptyDocument tests handling of minimal HTML
func TestURLExtractorEmptyDocument(t *testing.T) {
	html := `<html><head></head><body></body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	originalURL := "https://example.com/article/123"
	result := GenericUrlExtractor.Extract(doc.Selection, originalURL, []string{})

	// Should return original URL when no canonical links or meta tags
	if result.URL != originalURL {
		t.Errorf("Expected original URL %q, got %q", originalURL, result.URL)
	}

	if result.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got %q", result.Domain)
	}
}

// TestURLExtractorMalformedHTML tests robustness with bad HTML
func TestURLExtractorMalformedHTML(t *testing.T) {
	// Malformed HTML that parsers need to handle gracefully
	html := `
		<html
		<head>
			<link rel="canonical" href="https://news.example.com/story/123">
			<meta name="og:url" content="https://news.example.com/og/story/123"
		</head>
		<body>
			<h1>News Story</h1>
			<p>Content without closing tags
		</body>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := GenericUrlExtractor.Extract(doc.Selection, "https://news.example.com/ugly/story/123", []string{"og:url"})

	// Even with malformed HTML, should extract canonical link
	expectedURL := "https://news.example.com/story/123"
	if result.URL != expectedURL {
		t.Errorf("Expected canonical URL %q despite malformed HTML, got %q", expectedURL, result.URL)
	}
}

// TestURLExtractorPerformanceRealWorld tests performance with realistic content
func TestURLExtractorPerformanceRealWorld(t *testing.T) {
	// Large HTML document similar to real news sites
	html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>News Article</title>
			<link rel="canonical" href="https://news.example.com/2024/article">
			<meta name="og:url" content="https://news.example.com/2024/article">
			<!-- Many other meta tags -->
			<meta name="description" content="Article description">
			<meta name="keywords" content="news, breaking, important">
			<meta name="author" content="Reporter Name">
			<meta name="date" content="2024-01-19">
			<link rel="stylesheet" href="/css/main.css">
			<script src="/js/analytics.js"></script>
		</head>
		<body>
			<!-- Typical news site structure -->
			<header>
				<nav>
					<ul>
						<li><a href="/">Home</a></li>
						<li><a href="/politics">Politics</a></li>
						<li><a href="/tech">Technology</a></li>
					</ul>
				</nav>
			</header>
			<main>
				<article>
					<h1>Breaking News Article Title</h1>
					<div class="article-meta">
						<span class="author">By John Reporter</span>
						<time>January 19, 2024</time>
					</div>
					<div class="article-content">
						<p>First paragraph of content...</p>
						<p>Second paragraph...</p>
						<p>Many more paragraphs of content would be here...</p>
					</div>
				</article>
			</main>
			<aside>
				<div class="related-articles">
					<h3>Related Stories</h3>
					<!-- Related content -->
				</div>
			</aside>
			<footer>
				<!-- Footer content -->
			</footer>
		</body>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	// Performance test - should be fast even with large documents
	originalURL := "https://news.example.com/2024/article?utm_source=twitter"
	metaCache := []string{"og:url", "og:title", "twitter:url"}

	// Run multiple times to check consistency
	for i := 0; i < 10; i++ {
		result := GenericUrlExtractor.Extract(doc.Selection, originalURL, metaCache)
		
		expectedURL := "https://news.example.com/2024/article"
		if result.URL != expectedURL {
			t.Errorf("Iteration %d: Expected URL %q, got %q", i, expectedURL, result.URL)
		}
		
		if result.Domain != "news.example.com" {
			t.Errorf("Iteration %d: Expected domain 'news.example.com', got %q", i, result.Domain)
		}
	}
}

// BenchmarkURLExtractorRealWorld benchmarks with realistic HTML
func BenchmarkURLExtractorRealWorld(b *testing.B) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<link rel="canonical" href="https://example.com/article/123">
			<meta name="og:url" content="https://example.com/og/123">
			<title>Article Title</title>
		</head>
		<body>
			<article>
				<h1>Article Title</h1>
				<p>Content...</p>
			</article>
		</body>
		</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		b.Fatalf("Failed to parse HTML: %v", err)
	}

	originalURL := "https://example.com/ugly/article/123?ref=homepage"
	metaCache := []string{"og:url"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenericUrlExtractor.Extract(doc.Selection, originalURL, metaCache)
	}
}