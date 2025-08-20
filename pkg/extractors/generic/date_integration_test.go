// ABOUTME: End-to-end integration test for GenericDateExtractor with real-world HTML examples
// ABOUTME: Verifies complete date extraction pipeline works correctly with various content types

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestGenericDateExtractor_RealWorldIntegration(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		url         string
		metaCache   []string
		expected    string
		description string
	}{
		{
			name: "News article with meta tags",
			html: `<html>
				<head>
					<meta name="article:published_time" content="2023-12-01T15:30:00Z">
					<meta name="publishdate" content="2023-12-01">
				</head>
				<body>
					<article>
						<div class="entry-date">December 1, 2023</div>
						<p>News article content here...</p>
					</article>
				</body>
			</html>`,
			url:         "https://example.com/news/breaking-story",
			metaCache:   []string{"article:published_time", "publishdate"},
			expected:    "2023-12-01T15:30:00.000Z",
			description: "Should extract from highest priority meta tag",
		},
		{
			name: "Blog post with multiple date selectors",
			html: `<html>
				<head>
					<title>My Blog Post</title>
				</head>
				<body>
					<article class="hentry">
						<h1 class="entry-title">Blog Post Title</h1>
						<div class="meta">
							<span class="published">2023-12-01T09:15:00-08:00</span>
							<span class="author">John Doe</span>
						</div>
						<div class="entry-content">
							<p>Blog post content...</p>
						</div>
					</article>
				</body>
			</html>`,
			url:         "https://myblog.com/posts/my-post",
			metaCache:   []string{}, // No meta tags to match
			expected:    "2023-12-01T17:15:00.000Z", // Converted to UTC from PST
			description: "Should extract from CSS selectors and handle timezone conversion",
		},
		{
			name: "URL-based date extraction",
			html: `<html>
				<body>
					<h1>Article with Date in URL</h1>
					<p>This article has no explicit date metadata.</p>
				</body>
			</html>`,
			url:         "https://example.com/articles/2023/12/01/important-news",
			metaCache:   []string{}, // No meta tags
			expected:    "2023-12-01T08:00:00.000Z", // Parsed from URL in local timezone
			description: "Should extract date from URL when no metadata available",
		},
		{
			name: "Article with timestamp in milliseconds",
			html: `<html>
				<head>
					<meta name="pub_date" content="1701426600000">
				</head>
				<body>
					<h1>Timestamped Article</h1>
					<p>Article content...</p>
				</body>
			</html>`,
			url:         "https://example.com/timestamped-article",
			metaCache:   []string{"pub_date"},
			expected:    "2023-12-01T10:30:00.000Z",
			description: "Should parse millisecond timestamps correctly",
		},
		{
			name: "Priority test: Meta over selectors over URL",
			html: `<html>
				<head>
					<meta name="date" content="2023-12-01T10:00:00Z">
				</head>
				<body>
					<div class="entry-date">2023-12-02</div>
				</body>
			</html>`,
			url:         "https://example.com/2023/12/03/article",
			metaCache:   []string{"date"},
			expected:    "2023-12-01T10:00:00.000Z",
			description: "Should prioritize meta tags over selectors and URL",
		},
		{
			name: "Complex article with relative date text",
			html: `<html>
				<body>
					<article>
						<div class="byline">
							<span class="date">2 hours ago</span>
						</div>
						<p>Breaking news content...</p>
					</article>
				</body>
			</html>`,
			url:         "https://news.com/breaking",
			metaCache:   []string{},
			expected:    "", // We'll check this is within reasonable timeframe
			description: "Should parse relative dates like '2 hours ago'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			assert.NoError(t, err, "HTML parsing should not fail")

			result := GenericDateExtractor.Extract(doc.Selection, tt.url, tt.metaCache)

			if tt.expected == "" {
				// Special handling for relative dates - just verify it's a valid recent date
				assert.NotNil(t, result, "Should extract relative date")
				if result != nil {
					assert.Contains(t, *result, "202", "Should be a recent year (202x)")
					assert.Contains(t, *result, "T", "Should be ISO format")
					assert.Contains(t, *result, "Z", "Should be UTC format")
					t.Logf("Extracted relative date: %s", *result)
				}
			} else {
				assert.NotNil(t, result, "Date extraction should not return nil for test: %s", tt.description)
				if result != nil {
					assert.Equal(t, tt.expected, *result, "Extracted date should match expected value")
				}
			}

			t.Logf("Test '%s': %s -> Result: %v", tt.name, tt.description, result)
		})
	}
}

func TestGenericDateExtractor_EdgeCasesIntegration(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		url         string
		metaCache   []string
		shouldBeNil bool
		description string
	}{
		{
			name: "No date information anywhere",
			html: `<html>
				<body>
					<h1>Article with no date</h1>
					<p>Just content, no temporal information.</p>
				</body>
			</html>`,
			url:         "https://example.com/dateless-article",
			metaCache:   []string{},
			shouldBeNil: true,
			description: "Should return nil when no date information available",
		},
		{
			name: "Invalid date in meta tag",
			html: `<html>
				<head>
					<meta name="date" content="not a date">
				</head>
				<body>
					<p>Content with invalid date metadata.</p>
				</body>
			</html>`,
			url:         "https://example.com/invalid-date",
			metaCache:   []string{"date"},
			shouldBeNil: true,
			description: "Should return nil for unparseable date strings",
		},
		{
			name: "Multiple conflicting dates - meta cache filtering",
			html: `<html>
				<head>
					<meta name="date" content="2023-12-01">
					<meta name="pubdate" content="2023-12-02">
				</head>
				<body>
					<div class="entry-date">2023-12-03</div>
				</body>
			</html>`,
			url:         "https://example.com/article",
			metaCache:   []string{"pubdate"}, // Only pubdate in cache
			shouldBeNil: false,
			description: "Should extract from meta tags that match cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			assert.NoError(t, err)

			result := GenericDateExtractor.Extract(doc.Selection, tt.url, tt.metaCache)

			if tt.shouldBeNil {
				assert.Nil(t, result, "Should return nil: %s", tt.description)
			} else {
				assert.NotNil(t, result, "Should return valid date: %s", tt.description)
				if result != nil {
					// Basic validation that it's a proper ISO date
					assert.Contains(t, *result, "2023", "Should contain year")
					assert.Contains(t, *result, "T", "Should be ISO format")
					assert.Contains(t, *result, "Z", "Should be UTC timezone")
				}
			}

			t.Logf("Test '%s': %s -> Result: %v", tt.name, tt.description, result)
		})
	}
}

func TestGenericDateExtractor_JavaScriptCompatibilityIntegration(t *testing.T) {
	// Test cases that specifically verify JavaScript compatibility
	tests := []struct {
		name     string
		html     string
		url      string
		expected string
	}{
		{
			name: "Standard HTML meta with content attribute",
			html: `<html>
				<head>
					<meta name="article:published_time" content="2023-12-01T10:30:00Z">
				</head>
			</html>`,
			url:      "https://example.com",
			expected: "2023-12-01T10:30:00.000Z",
		},
		{
			name: "Non-standard meta with value attribute (old fixture format)",
			html: `<html>
				<head>
					<meta name="pubdate" value="2023-12-01T10:30:00Z">
				</head>
			</html>`,
			url:      "https://example.com",
			expected: "2023-12-01T10:30:00.000Z",
		},
		{
			name: "Millisecond timestamp (JavaScript compatible)",
			html: `<html>
				<head>
					<meta name="pubdate" content="1701426600000">
				</head>
			</html>`,
			url:      "https://example.com",
			expected: "2023-12-01T10:30:00.000Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			assert.NoError(t, err)

			// Create appropriate meta cache
			var metaCache []string
			doc.Find("meta").Each(func(i int, s *goquery.Selection) {
				if name, exists := s.Attr("name"); exists {
					metaCache = append(metaCache, name)
				}
			})

			result := GenericDateExtractor.Extract(doc.Selection, tt.url, metaCache)

			assert.NotNil(t, result, "JavaScript-compatible extraction should work")
			assert.Equal(t, tt.expected, *result, "Should match JavaScript behavior exactly")
		})
	}
}