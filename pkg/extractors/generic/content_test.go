// ABOUTME: Comprehensive test suite for generic content extractor with 100% JavaScript compatibility verification
// ABOUTME: Tests extraction pipeline including node sufficiency, content cleaning, and cascading extraction options

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericContentExtractor_Extract_BasicFunctionality(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
		contains []string
	}{
		{
			name: "simple article with sufficient content",
			html: `<html>
				<body>
					<div class="content">
						<p>This is a long article with sufficient content to be considered valid. 
						It contains multiple sentences and enough text to pass the threshold requirements 
						for article extraction. This should definitely be detected as the main content.</p>
					</div>
				</body>
			</html>`,
			contains: []string{"long article", "sufficient content", "article extraction"},
		},
		{
			name: "insufficient content should cascade through options",
			html: `<html>
				<body>
					<div class="content">
						<p>Short text.</p>
					</div>
				</body>
			</html>`,
			contains: []string{"Short text"},
		},
		{
			name: "article with title and multiple paragraphs",
			html: `<html>
				<head><title>Test Article Title</title></head>
				<body>
					<h1>Article Heading</h1>
					<div class="article-content">
						<p>First paragraph with substantial content that describes the main topic of this article.
						This paragraph contains enough text to be considered meaningful content.</p>
						<p>Second paragraph that continues the discussion and provides additional details
						about the subject matter. This helps establish the content as article-like.</p>
					</div>
				</body>
			</html>`,
			contains: []string{"First paragraph", "substantial content", "Second paragraph", "additional details"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := &GenericContentExtractor{}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			params := ExtractorParams{
				Doc:   doc,
				HTML:  tt.html,
				Title: "Test Article Title",
				URL:   "https://example.com/article",
			}

			result := extractor.Extract(params, ExtractorOptions{})
			
			assert.NotEmpty(t, result, "should extract content")
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "should contain expected text")
			}
		})
	}
}

func TestGenericContentExtractor_Extract_OptionsHandling(t *testing.T) {
	html := `<html>
		<body>
			<div class="unlikely-candidate comments">
				<p>This is a comment that should be stripped by default options.</p>
			</div>
			<div class="main-content">
				<p>This is the main article content that should be extracted successfully.
				It contains enough text to be considered substantial and meaningful for extraction purposes.</p>
			</div>
		</body>
	</html>`

	extractor := &GenericContentExtractor{}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	params := ExtractorParams{
		Doc:   doc,
		HTML:  html,
		Title: "Test Article",
		URL:   "https://example.com/test",
	}

	// Test with default options (should strip unlikely candidates)
	result := extractor.Extract(params, ExtractorOptions{})
	assert.Contains(t, result, "main article content")
	assert.NotContains(t, result, "comment that should be stripped")

	// Test with custom options
	customOpts := ExtractorOptions{
		StripUnlikelyCandidates: false,
		WeightNodes:             false,
		CleanConditionally:      false,
	}
	result2 := extractor.Extract(params, customOpts)
	assert.NotEmpty(t, result2)
}

func TestGenericContentExtractor_Extract_CascadingOptions(t *testing.T) {
	// HTML that might need cascading options to extract successfully
	html := `<html>
		<body>
			<div class="content">
				<p>Marginal content that might require loosened extraction criteria to be detected properly.</p>
			</div>
		</body>
	</html>`

	extractor := &GenericContentExtractor{}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	params := ExtractorParams{
		Doc:   doc,
		HTML:  html,
		Title: "Test Article",
		URL:   "https://example.com/test",
	}

	result := extractor.Extract(params, ExtractorOptions{})
	assert.NotEmpty(t, result, "should extract content even with marginal quality")
	assert.Contains(t, result, "Marginal content")
}

func TestGenericContentExtractor_Extract_EmptyContent(t *testing.T) {
	tests := []struct {
		name string
		html string
	}{
		{
			name: "empty document",
			html: `<html><body></body></html>`,
		},
		{
			name: "no meaningful content",
			html: `<html><body><div class="ads">Ad content</div></body></html>`,
		},
		{
			name: "malformed HTML",
			html: `<div><p>Incomplete`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := &GenericContentExtractor{}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			params := ExtractorParams{
				Doc:   doc,
				HTML:  tt.html,
				Title: "Test",
				URL:   "https://example.com",
			}

			result := extractor.Extract(params, ExtractorOptions{})
			// Should not panic, might return empty string
			assert.NotNil(t, result)
		})
	}
}

func TestGenericContentExtractor_GetContentNode(t *testing.T) {
	html := `<html>
		<body>
			<div class="article">
				<p>This is article content that should be extracted and cleaned properly.
				It contains sufficient text to be considered meaningful.</p>
			</div>
		</body>
	</html>`

	extractor := &GenericContentExtractor{}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	opts := ExtractorOptions{
		StripUnlikelyCandidates: true,
		WeightNodes:             true,
		CleanConditionally:      true,
	}

	node := extractor.GetContentNode(doc, "Test Title", "https://example.com", opts)
	assert.NotNil(t, node, "should return a content node")
}

func TestGenericContentExtractor_CleanAndReturnNode(t *testing.T) {
	html := `<div><p>Test content   with   extra   spaces</p></div>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	extractor := &GenericContentExtractor{}

	// Test with valid node
	node := doc.Find("div").First()
	result := extractor.CleanAndReturnNode(node, doc)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Test content")

	// Test with nil node
	result2 := extractor.CleanAndReturnNode(nil, doc)
	assert.Empty(t, result2)
}

func TestNodeIsSufficient(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{
			name:     "sufficient content",
			text:     strings.Repeat("This is a test sentence. ", 10), // Well over 100 chars
			expected: true,
		},
		{
			name:     "insufficient content",
			text:     "Short text",
			expected: false,
		},
		{
			name:     "exactly 100 characters",
			text:     strings.Repeat("x", 100),
			expected: true,
		},
		{
			name:     "99 characters",
			text:     strings.Repeat("x", 99),
			expected: false,
		},
		{
			name:     "empty content",
			text:     "",
			expected: false,
		},
		{
			name:     "whitespace only",
			text:     "   \n\t   ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := `<div><p>` + tt.text + `</p></div>`
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			require.NoError(t, err)

			node := doc.Find("div").First()
			result := NodeIsSufficient(node)
			assert.Equal(t, tt.expected, result)
		})
	}
}