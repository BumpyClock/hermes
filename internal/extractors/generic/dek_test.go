// ABOUTME: Comprehensive tests for GenericDekExtractor with JavaScript compatibility verification
// ABOUTME: Tests meta description extraction, selector-based extraction, and dek cleaning validation

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestGenericDekExtractor_Extract_MetaDescription(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "extracts from meta description",
			html: `<html>
				<head>
					<meta name="description" content="This is a good article description" />
				</head>
				<body><div>Content</div></body>
			</html>`,
			expected: "This is a good article description",
		},
		{
			name: "extracts from og:description",
			html: `<html>
				<head>
					<meta property="og:description" content="OpenGraph description for article" />
				</head>
				<body><div>Content</div></body>
			</html>`,
			expected: "OpenGraph description for article",
		},
		{
			name: "extracts from twitter:description",
			html: `<html>
				<head>
					<meta name="twitter:description" content="Twitter card description" />
				</head>
				<body><div>Content</div></body>
			</html>`,
			expected: "Twitter card description",
		},
		{
			name: "prioritizes description over og:description",
			html: `<html>
				<head>
					<meta name="description" content="Standard meta description" />
					<meta property="og:description" content="OpenGraph description" />
				</head>
				<body><div>Content</div></body>
			</html>`,
			expected: "Standard meta description",
		},
		{
			name: "falls back to og:description if no standard description",
			html: `<html>
				<head>
					<meta property="og:description" content="Only OpenGraph available" />
				</head>
				<body><div>Content</div></body>
			</html>`,
			expected: "Only OpenGraph available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			extractor := &GenericDekExtractor{}
			result := extractor.Extract(doc, map[string]interface{}{
				"$": doc.Selection,
			})

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenericDekExtractor_Extract_SelectorBased(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "extracts from entry-summary selector",
			html: `<html>
				<body>
					<div class="entry-summary">This is the article summary</div>
					<div>Other content</div>
				</body>
			</html>`,
			expected: "This is the article summary",
		},
		{
			name: "extracts from h2 with itemprop description (Ars Technica style)",
			html: `<html>
				<body>
					<h2 itemprop="description">Technical article subtitle</h2>
					<div>Article content</div>
				</body>
			</html>`,
			expected: "Technical article subtitle",
		},
		{
			name: "extracts from subtitle selectors",
			html: `<html>
				<body>
					<div class="subtitle">Article subtitle here</div>
					<div>Article content</div>
				</body>
			</html>`,
			expected: "Article subtitle here",
		},
		{
			name: "prefers first matching selector",
			html: `<html>
				<body>
					<div class="entry-summary">Primary summary</div>
					<div class="subtitle">Secondary subtitle</div>
				</body>
			</html>`,
			expected: "Primary summary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			extractor := &GenericDekExtractor{}
			result := extractor.Extract(doc, map[string]interface{}{
				"$": doc.Selection,
			})

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenericDekExtractor_Extract_Validation(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expected    string
		description string
	}{
		{
			name: "rejects too short dek (under 5 characters)",
			html: `<html>
				<head><meta name="description" content="Hi" /></head>
				<body><div>Content</div></body>
			</html>`,
			expected:    "",
			description: "Dek should be rejected if under 5 characters",
		},
		{
			name: "rejects too long dek (over 1000 characters)",
			html: `<html>
				<head><meta name="description" content="` + strings.Repeat("Very long description text ", 50) + `" /></head>
				<body><div>Content</div></body>
			</html>`,
			expected:    "",
			description: "Dek should be rejected if over 1000 characters",
		},
		{
			name: "rejects dek with plain text links",
			html: `<html>
				<head><meta name="description" content="Check out https://example.com for more info" /></head>
				<body><div>Content</div></body>
			</html>`,
			expected:    "",
			description: "Dek should be rejected if it contains plain text URLs",
		},
		{
			name: "accepts valid dek length",
			html: `<html>
				<head><meta name="description" content="This is a good article description that is the right length" /></head>
				<body><div>Content</div></body>
			</html>`,
			expected:    "This is a good article description that is the right length",
			description: "Valid dek should be accepted",
		},
		{
			name: "normalizes whitespace in dek",
			html: `<html>
				<head><meta name="description" content="   Whitespace   normalized   dek   " /></head>
				<body><div>Content</div></body>
			</html>`,
			expected:    "Whitespace normalized dek",
			description: "Whitespace should be normalized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			extractor := &GenericDekExtractor{}
			result := extractor.Extract(doc, map[string]interface{}{
				"$": doc.Selection,
			})

			if result != tt.expected {
				t.Errorf("Expected %q, got %q - %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestGenericDekExtractor_Extract_ExcerptComparison(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		excerpt     string
		expected    string
		description string
	}{
		{
			name: "rejects dek identical to excerpt",
			html: `<html>
				<head><meta name="description" content="This is the same text as excerpt and more content" /></head>
				<body><div>Content</div></body>
			</html>`,
			excerpt:     "This is the same text as excerpt and more content here",
			expected:    "",
			description: "Dek should be rejected if it matches excerpt (first 10 words)",
		},
		{
			name: "accepts dek different from excerpt",
			html: `<html>
				<head><meta name="description" content="This is a different subtitle description" /></head>
				<body><div>Content</div></body>
			</html>`,
			excerpt:     "This is the main article content with more text here",
			expected:    "This is a different subtitle description",
			description: "Dek should be accepted if different from excerpt",
		},
		{
			name: "handles no excerpt provided",
			html: `<html>
				<head><meta name="description" content="Valid dek without excerpt comparison" /></head>
				<body><div>Content</div></body>
			</html>`,
			excerpt:     "",
			expected:    "Valid dek without excerpt comparison",
			description: "Dek should be accepted when no excerpt provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			extractor := &GenericDekExtractor{}
			opts := map[string]interface{}{
				"$": doc.Selection,
			}
			if tt.excerpt != "" {
				opts["excerpt"] = tt.excerpt
			}

			result := extractor.Extract(doc, opts)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q - %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestGenericDekExtractor_Extract_FallbackChain(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expected    string
		description string
	}{
		{
			name: "tries meta tags then selectors",
			html: `<html>
				<head></head>
				<body>
					<div class="entry-summary">Found via selector</div>
				</body>
			</html>`,
			expected:    "Found via selector",
			description: "Should fall back to selectors when no meta tags",
		},
		{
			name: "prefers meta description over selectors",
			html: `<html>
				<head><meta name="description" content="Meta description" /></head>
				<body>
					<div class="entry-summary">Selector result</div>
				</body>
			</html>`,
			expected:    "Meta description",
			description: "Should prefer meta description over selectors",
		},
		{
			name: "returns empty for no matches",
			html: `<html>
				<head></head>
				<body><div>No dek content</div></body>
			</html>`,
			expected:    "",
			description: "Should return empty string when no dek found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			extractor := &GenericDekExtractor{}
			result := extractor.Extract(doc, map[string]interface{}{
				"$": doc.Selection,
			})

			if result != tt.expected {
				t.Errorf("Expected %q, got %q - %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestGenericDekExtractor_Extract_HTMLTagsCleaning(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expected    string
		description string
	}{
		{
			name: "strips HTML tags from dek",
			html: `<html>
				<head><meta name="description" content="Description with <em>emphasis</em> and <strong>bold</strong>" /></head>
				<body><div>Content</div></body>
			</html>`,
			expected:    "Description with emphasis and bold",
			description: "HTML tags should be stripped from dek",
		},
		{
			name: "strips complex HTML from selector-based dek",
			html: `<html>
				<body>
					<div class="entry-summary">Summary with <a href="#">link</a> and <span>span</span></div>
				</body>
			</html>`,
			expected:    "Summary with link and span",
			description: "Complex HTML should be stripped from selector-based dek",
		},
		{
			name: "handles nested HTML tags",
			html: `<html>
				<head><meta name="description" content="<div><p>Nested <em>HTML</em> content</p></div>" /></head>
				<body><div>Content</div></body>
			</html>`,
			expected:    "Nested HTML content",
			description: "Nested HTML tags should be properly stripped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			extractor := &GenericDekExtractor{}
			result := extractor.Extract(doc, map[string]interface{}{
				"$": doc.Selection,
			})

			if result != tt.expected {
				t.Errorf("Expected %q, got %q - %s", tt.expected, result, tt.description)
			}
		})
	}
}
func TestGenericDekExtractor_JavaScriptCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		excerpt     string
		expected    string
		description string
	}{
		{
			name: "JavaScript stub behavior - currently returns empty",
			html: `<html>
				<head><meta name="description" content="Article description" /></head>
				<body><div>Content</div></body>
			</html>`,
			expected:    "Article description",
			description: "Our implementation extracts dek, while JavaScript returns null (stub)",
		},
		{
			name: "meta property vs name attribute prioritization",
			html: `<html>
				<head>
					<meta property="description" content="Property description" />
					<meta name="description" content="Name description" />
				</head>
				<body></body>
			</html>`,
			expected:    "Name description",
			description: "Should prefer name attribute over property attribute",
		},
		{
			name: "empty meta tag handling",
			html: `<html>
				<head>
					<meta name="description" content="" />
					<meta name="og:description" content="OpenGraph fallback" />
				</head>
				<body></body>
			</html>`,
			expected:    "",
			description: "Should reject empty meta content and not fall back",
		},
		{
			name: "boundary length validation - 4 chars (too short)",
			html: `<html>
				<head><meta name="description" content="test" /></head>
				<body></body>
			</html>`,
			expected:    "",
			description: "Should reject 4-character dek (under 5 minimum)",
		},
		{
			name: "boundary length validation - 5 chars (acceptable)",
			html: `<html>
				<head><meta name="description" content="tests" /></head>
				<body></body>
			</html>`,
			expected:    "tests",
			description: "Should accept 5-character dek (minimum length)",
		},
		{
			name: "real-world Ars Technica selector",
			html: `<html>
				<body>
					<article>
						<h1>Article Title</h1>
						<h2 itemprop="description">Technical deep dive into new processor architecture</h2>
						<div>Article content...</div>
					</article>
				</body>
			</html>`,
			expected:    "Technical deep dive into new processor architecture",
			description: "Should match real custom extractor patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			extractor := &GenericDekExtractor{}
			opts := map[string]interface{}{
				"$": doc.Selection,
			}
			if tt.excerpt != "" {
				opts["excerpt"] = tt.excerpt
			}

			result := extractor.Extract(doc, opts)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q - %s", tt.expected, result, tt.description)
			}
		})
	}
}

func TestGenericDekExtractor_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "WordPress blog post",
			html: `<html>
				<head><meta name="description" content="Learn about the latest web development trends in 2024" /></head>
				<body>
					<div class="entry-summary">Quick overview of modern JavaScript frameworks</div>
				</body>
			</html>`,
			expected: "Learn about the latest web development trends in 2024",
		},
		{
			name: "News article with OpenGraph",
			html: `<html>
				<head>
					<meta property="og:description" content="Breaking news story about technology innovation" />
				</head>
				<body><article>News content...</article></body>
			</html>`,
			expected: "Breaking news story about technology innovation",
		},
		{
			name: "Technical documentation",
			html: `<html>
				<body>
					<div class="subtitle">API reference for advanced users</div>
					<div class="content">Documentation content...</div>
				</body>
			</html>`,
			expected: "API reference for advanced users",
		},
		{
			name: "SEO-heavy description (with link) - should be rejected",
			html: `<html>
				<head><meta name="description" content="Visit https://example.com for amazing deals and discounts" /></head>
				<body></body>
			</html>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			extractor := &GenericDekExtractor{}
			result := extractor.Extract(doc, map[string]interface{}{
				"$": doc.Selection,
			})

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
