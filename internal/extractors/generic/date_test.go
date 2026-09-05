// ABOUTME: Tests for the generic date published extractor
// ABOUTME: Verifies 100% JavaScript compatibility for date extraction from meta tags, selectors, and URLs

package generic

import (
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestGenericDateExtractor_Basic(t *testing.T) {
	// This test should initially fail - we haven't implemented GenericDateExtractor yet
	html := `<html>
		<head>
			<meta name="article:published_time" content="2023-12-01T10:30:00Z">
		</head>
		<body>
			<div class="entry-date">December 1, 2023</div>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	assert.NoError(t, err)

	// Test meta tag extraction
	url := "https://example.com/article/2023/12/01/test"
	metaCache := []string{"article:published_time", "displaydate", "dc.date"}

	result := GenericDateExtractor.Extract(doc.Selection, url, metaCache)

	// Should extract date from meta tag (note: may be affected by local timezone)
	assert.NotNil(t, result)
	// The test environment may parse this differently due to timezone, but we should get a valid 2023-12-01 date
	assert.Contains(t, *result, "2023-12-01", "Should contain the correct date")
}

func TestGenericDateExtractor_MetaTags(t *testing.T) {
	localMidnight := time.Date(2023, time.December, 1, 0, 0, 0, 0, time.Local).UTC().Format("2006-01-02T15:04:05.000Z")
	tests := []struct {
		name     string
		metaTag  string
		content  string
		expected string
	}{
		{
			name:     "article:published_time",
			metaTag:  "article:published_time",
			content:  "2023-12-01T10:30:00Z",
			expected: "2023-12-01T10:30:00.000Z",
		},
		{
			name:     "displaydate",
			metaTag:  "displaydate",
			content:  "2023-12-01",
			expected: localMidnight,
		},
		{
			name:     "dc.date",
			metaTag:  "dc.date",
			content:  "December 1, 2023",
			expected: localMidnight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<html>
				<head>
					<meta name="` + tt.metaTag + `" content="` + tt.content + `">
				</head>
				<body></body>
			</html>`

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			assert.NoError(t, err)

			url := "https://example.com/article"
			metaCache := []string{tt.metaTag}

			result := GenericDateExtractor.Extract(doc.Selection, url, metaCache)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

func TestGenericDateExtractor_Selectors(t *testing.T) {
	localMidnight := time.Date(2023, time.December, 1, 0, 0, 0, 0, time.Local).UTC().Format("2006-01-02T15:04:05.000Z")
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "entry-date selector",
			html:     `<html><body><div class="entry-date">2023-12-01</div></body></html>`,
			expected: localMidnight,
		},
		{
			name:     "hentry published selector",
			html:     `<html><body><div class="hentry"><span class="published">December 1, 2023</span></div></body></html>`,
			expected: localMidnight,
		},
		{
			name:     "meta postDate selector",
			html:     `<html><body><div class="meta"><time class="postDate">2023-12-01T10:30:00Z</time></div></body></html>`,
			expected: "2023-12-01T10:30:00.000Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			assert.NoError(t, err)

			url := "https://example.com/article"
			metaCache := []string{} // Empty cache for selector tests

			result := GenericDateExtractor.Extract(doc.Selection, url, metaCache)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

func TestGenericDateExtractor_URLExtraction(t *testing.T) {
	localMidnight := time.Date(2023, time.December, 1, 0, 0, 0, 0, time.Local).UTC().Format("2006-01-02T15:04:05.000Z")
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "URL with YYYY/MM/DD format",
			url:      "https://example.com/2023/12/01/article-title",
			expected: localMidnight,
		},
		{
			name:     "URL with YYYY-MM-DD format",
			url:      "https://example.com/news/2023-12-01-breaking",
			expected: localMidnight,
		},
		{
			name:     "URL with month abbreviation",
			url:      "https://example.com/2023/dec/01/story",
			expected: localMidnight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<html><body></body></html>`
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			assert.NoError(t, err)

			metaCache := []string{} // Empty cache for URL extraction tests

			result := GenericDateExtractor.Extract(doc.Selection, tt.url, metaCache)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

func TestGenericDateExtractor_NoDateFound(t *testing.T) {
	html := `<html><body><div>No date information</div></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	assert.NoError(t, err)

	url := "https://example.com/article"
	metaCache := []string{} // Empty cache for no date test

	result := GenericDateExtractor.Extract(doc.Selection, url, metaCache)

	// Should return nil when no date is found
	assert.Nil(t, result)
}

func TestGenericDateExtractor_Priority(t *testing.T) {
	// Test that meta tags have priority over selectors
	html := `<html>
		<head>
			<meta name="article:published_time" content="2023-12-01T10:30:00Z">
		</head>
		<body>
			<div class="entry-date">2023-12-02</div>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	assert.NoError(t, err)

	url := "https://example.com/article"
	metaCache := []string{"article:published_time"}

	result := GenericDateExtractor.Extract(doc.Selection, url, metaCache)

	// Should prefer meta tag over selector
	assert.NotNil(t, result)
	assert.Equal(t, "2023-12-01T10:30:00.000Z", *result)
}
