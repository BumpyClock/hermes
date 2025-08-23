// ABOUTME: Comprehensive test suite for description extraction functionality
// ABOUTME: Tests meta tag extraction, JSON-LD parsing, and validation strategies

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/BumpyClock/hermes/pkg/resource"
)

func TestGenericDescriptionExtractor_Extract(t *testing.T) {
	tests := []struct {
		name      string
		html      string
		url       string
		metaCache []string
		expected  string
	}{
		{
			name: "extracts from standard meta description",
			html: `<html>
				<head>
					<meta name="description" content="This is a comprehensive tech news website." />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "This is a comprehensive tech news website.",
		},
		{
			name: "extracts from og:description",
			html: `<html>
				<head>
					<meta property="og:description" content="Breaking technology news and analysis." />
				</head>
			</html>`,
			url:       "https://techsite.com",
			metaCache: []string{},
			expected:  "Breaking technology news and analysis.",
		},
		{
			name: "extracts from twitter:description",
			html: `<html>
				<head>
					<meta name="twitter:description" content="Your source for startup and technology news." />
				</head>
			</html>`,
			url:       "https://startup.com",
			metaCache: []string{},
			expected:  "Your source for startup and technology news.",
		},
		{
			name: "prefers standard description over og:description",
			html: `<html>
				<head>
					<meta name="description" content="The official site description" />
					<meta property="og:description" content="Social media description" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "The official site description",
		},
		{
			name: "extracts from JSON-LD WebSite",
			html: `<html>
				<head>
					<script type="application/ld+json">
					{
						"@context": "https://schema.org",
						"@type": "WebSite",
						"name": "Tech News",
						"description": "The latest technology news and reviews."
					}
					</script>
				</head>
			</html>`,
			url:       "https://technews.com",
			metaCache: []string{},
			expected:  "The latest technology news and reviews.",
		},
		{
			name: "extracts from JSON-LD NewsMediaOrganization",
			html: `<html>
				<head>
					<script type="application/ld+json">
					{
						"@context": "https://schema.org",
						"@type": "NewsMediaOrganization",
						"name": "News Site",
						"description": "Independent journalism with a mission to inform."
					}
					</script>
				</head>
			</html>`,
			url:       "https://newssite.com",
			metaCache: []string{},
			expected:  "Independent journalism with a mission to inform.",
		},
		{
			name: "extracts from JSON-LD Organization",
			html: `<html>
				<head>
					<script type="application/ld+json">
					{
						"@context": "https://schema.org",
						"@type": "Organization",
						"name": "Media Company",
						"description": "Creating quality content for digital audiences."
					}
					</script>
				</head>
			</html>`,
			url:       "https://media.com",
			metaCache: []string{},
			expected:  "Creating quality content for digital audiences.",
		},
		{
			name: "handles nested JSON-LD publisher description",
			html: `<html>
				<head>
					<script type="application/ld+json">
					{
						"@context": "https://schema.org",
						"@type": "Article",
						"headline": "Test Article",
						"publisher": {
							"@type": "Organization",
							"name": "Publisher",
							"description": "Publisher's site description"
						}
					}
					</script>
				</head>
			</html>`,
			url:       "https://publisher.com",
			metaCache: []string{},
			expected:  "Publisher's site description",
		},
		{
			name: "rejects too short descriptions",
			html: `<html>
				<head>
					<meta name="description" content="Short" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "",
		},
		{
			name: "rejects descriptions with URLs",
			html: `<html>
				<head>
					<meta name="description" content="Visit https://example.com for more information" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "",
		},
		{
			name: "rejects article-specific descriptions",
			html: `<html>
				<head>
					<meta name="description" content="In this article, we explore the latest trends" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "",
		},
		{
			name: "handles invalid JSON-LD gracefully",
			html: `<html>
				<head>
					<script type="application/ld+json">
					{ invalid json }
					</script>
					<meta name="description" content="Valid site description from meta tag" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "Valid site description from meta tag",
		},
		{
			name: "empty when no valid data",
			html: `<html>
				<head>
					<meta name="other" content="not relevant" />
				</head>
			</html>`,
			url:       "https://example.com",
			metaCache: []string{},
			expected:  "",
		},
		{
			name: "NPR-style description",
			html: `<html>
				<head>
					<meta name="description" content="Top stories in the U.S. and world news, politics, health, science, business, music, arts and culture. Nonprofit journalism with a mission. This is NPR." />
				</head>
			</html>`,
			url:       "https://www.npr.org",
			metaCache: []string{},
			expected:  "Top stories in the U.S. and world news, politics, health, science, business, music, arts and culture. Nonprofit journalism with a mission. This is NPR.",
		},
		{
			name: "The Verge-style description",
			html: `<html>
				<head>
					<meta name="description" content="The Verge is about technology and how it makes us feel. Founded in 2011, we offer our audience everything from breaking news to reviews to award-winning features and investigations, on our site, in video, and in podcasts." />
				</head>
			</html>`,
			url:       "https://www.theverge.com",
			metaCache: []string{},
			expected:  "The Verge is about technology and how it makes us feel. Founded in 2011, we offer our audience everything from breaking news to reviews to award-winning features and investigations, on our site, in video, and in podcasts.",
		},
	}

	extractor := &GenericDescriptionExtractor{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Apply same normalization as in real extraction pipeline
			doc = resource.NormalizeMetaTags(doc)

			result := extractor.Extract(doc.Selection, tt.url, tt.metaCache)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenericDescriptionExtractor_IsValidDescription(t *testing.T) {
	extractor := &GenericDescriptionExtractor{}

	tests := []struct {
		input    string
		expected bool
	}{
		{"Valid site description for testing purposes", true},
		{"", false},                                    // empty
		{"Short", false},                              // too short
		{"This is a very long description that goes on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and on and exceeds the reasonable limit", false}, // too long
		{"Visit https://example.com for more info", false}, // contains URL
		{"Check out http://site.com", false},               // contains URL
		{"In this article we discuss topics", false},       // article-specific
		{"This article covers many subjects", false},       // article-specific
		{"Read more about this topic", false},              // article-specific
		{"Continue reading for details", false},            // article-specific
		{"Full story: Latest news update", false},          // article-specific
		{"Technology news and analysis site", true},        // valid
		{"Independent journalism with a mission", true},    // valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractor.isValidDescription(tt.input)
			if result != tt.expected {
				t.Errorf("For input %q, expected %v, got %v", tt.input, tt.expected, result)
			}
		})
	}
}

func TestGenericDescriptionExtractor_CleanDescription(t *testing.T) {
	extractor := &GenericDescriptionExtractor{}

	tests := []struct {
		input    string
		expected string
	}{
		{"Normal description", "Normal description"},
		{"  Spaced  description  ", "Spaced description"},
		{"Description with\nmultiple\nlines", "Description with multiple lines"},
		{"Description - Read more", "Description"},
		{"Description | Read more", "Description"},
		{"Description - Continue reading", "Description"},
		{"Description | Continue reading", "Description"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractor.cleanDescription(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenericDescriptionExtractor_ExtractFromMetaTags(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "standard meta name description",
			html: `<html><head><meta name="description" content="Site description" /></head></html>`,
			expected: "Site description",
		},
		{
			name: "og:description property",
			html: `<html><head><meta property="og:description" content="OG description" /></head></html>`,
			expected: "OG description",
		},
		{
			name: "twitter:description name",
			html: `<html><head><meta name="twitter:description" content="Twitter description" /></head></html>`,
			expected: "Twitter description",
		},
		{
			name: "dc.description",
			html: `<html><head><meta name="dc.description" content="Dublin Core description" /></head></html>`,
			expected: "Dublin Core description",
		},
		{
			name: "no valid description",
			html: `<html><head><meta name="other" content="Not a description" /></head></html>`,
			expected: "",
		},
	}

	extractor := &GenericDescriptionExtractor{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Apply same normalization as in real extraction pipeline
			doc = resource.NormalizeMetaTags(doc)

			result := extractor.extractFromMetaTags(doc.Selection)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}