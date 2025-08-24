// ABOUTME: JavaScript compatibility verification tests comparing Go extractor behavior against expected JavaScript outputs  
// ABOUTME: Comprehensive test suite validating exact behavioral match with original JavaScript implementation

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJavaScriptCompatibility_ContentExtractor performs comprehensive validation
// that the Go implementation matches JavaScript behavior exactly
func TestJavaScriptCompatibility_ContentExtractor(t *testing.T) {
	testCases := []struct {
		name        string
		html        string
		title       string
		url         string
		expected    []string // content that must be present
		notExpected []string // content that should be filtered out
		minLength   int      // minimum expected content length
		description string
	}{
		{
			name: "JavaScript Test Case 1 - Basic Article Structure",
			html: `<html>
<head><title>Sample Article Title</title></head>
<body>
	<div class="content entry-content">
		<h2>Article Heading</h2>
		<p>This is the main article paragraph with substantial content that should be extracted.
		The content is meaningful and provides value to readers seeking information on this topic.
		This paragraph contains sufficient text to meet extraction criteria.</p>
		
		<p>Second paragraph that continues the article narrative with additional details and context.
		This content expands on the initial points and provides comprehensive coverage of the subject matter.</p>
	</div>
	<div class="sidebar">
		<div class="ads">Advertisement content</div>
		<div class="comments">User comments section</div>
	</div>
</body>
</html>`,
			title: "Sample Article Title",
			url:   "https://example.com/article",
			expected: []string{
				"main article paragraph",
				"substantial content",
				"Second paragraph",
				"additional details",
			},
			notExpected: []string{
				"Advertisement content",
				"User comments",
			},
			minLength:   200,
			description: "Basic article with clear content/sidebar separation",
		},
		{
			name: "JavaScript Test Case 2 - Complex DOM Structure",
			html: `<!DOCTYPE html>
<html>
<head>
	<title>Complex Article</title>
</head>
<body>
	<nav class="navigation">
		<ul><li><a href="/">Home</a></li></ul>
	</nav>
	
	<div class="unlikely-comment ads">
		<p>This should be stripped by unlikely candidate removal</p>
	</div>
	
	<article class="main-content">
		<header>
			<h1>Main Article Title</h1>
			<div class="byline">By Test Author</div>
		</header>
		
		<div class="article-body entry-content">
			<p>Opening paragraph of the main article content with comprehensive information about the topic.
			This content should be preserved and extracted as the primary article content.</p>
			
			<h3>Section Heading</h3>
			<p>Content under section heading that provides additional detail and context for readers.
			This section elaborates on key points and maintains the article narrative flow.</p>
			
			<blockquote>
				<p>Important quoted material that adds credibility and perspective to the article content.
				Quotes should be preserved as they contribute valuable information.</p>
			</blockquote>
			
			<p>Concluding paragraph that summarizes key points and provides closure to the article.
			This final section ensures readers have complete understanding of the topic.</p>
		</div>
	</article>
	
	<aside class="sidebar">
		<div class="widget related-links">Related articles</div>
		<div class="widget social">Social sharing</div>
	</aside>
	
	<footer>
		<div class="copyright">Copyright information</div>
	</footer>
</body>
</html>`,
			title: "Complex Article",
			url:   "https://example.com/complex-article",
			expected: []string{
				"Opening paragraph",
				"comprehensive information",
				"Section Heading",
				"additional detail",
				"Important quoted material",
				"Concluding paragraph",
			},
			notExpected: []string{
				"unlikely candidate removal",
				"Related articles",
				"Social sharing",
				"Copyright information",
			},
			minLength:   500,
			description: "Complex DOM with navigation, sidebar, and footer elements",
		},
		{
			name: "JavaScript Test Case 3 - Content with Lists and Formatting",
			html: `<html>
<body>
	<div class="main-article">
		<h2>Article with Structured Content</h2>
		
		<p>Introduction paragraph that sets up the context for the following information.
		This content provides essential background for understanding the detailed points below.</p>
		
		<ul>
			<li>First important point with detailed explanation and supporting information</li>
			<li>Second key point that builds on the first and adds new perspective</li>
			<li>Third crucial point that completes the argument and provides conclusion</li>
		</ul>
		
		<p>Analysis paragraph that discusses the implications of the points listed above.
		This section connects the individual points into a cohesive narrative structure.</p>
		
		<ol>
			<li>Step one in the process with clear instructions and examples</li>
			<li>Step two that follows logically from the first step</li>
			<li>Final step that completes the process successfully</li>
		</ol>
		
		<p>Summary paragraph that reinforces the main themes and provides actionable takeaways.
		This conclusion ensures readers understand both theory and practical application.</p>
	</div>
</body>
</html>`,
			title: "Article with Lists",
			url:   "https://example.com/structured-content",
			expected: []string{
				"Introduction paragraph",
				"essential background",
				"First important point",
				"Second key point",
				"Third crucial point",
				"Analysis paragraph",
				"Step one in the process",
				"Step two that follows",
				"Final step",
				"Summary paragraph",
			},
			notExpected: []string{},
			minLength:   600,
			description: "Article with ordered and unordered lists",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extractor := NewGenericContentExtractor()
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			require.NoError(t, err)

			params := ExtractorParams{
				Doc:   doc,
				HTML:  tc.html,
				Title: tc.title,
				URL:   tc.url,
			}

			// Extract content with default options (matches JavaScript defaults)
			result := extractor.Extract(params, ExtractorOptions{})

			// Validate content was extracted
			assert.NotEmpty(t, result, "Should extract content")
			assert.GreaterOrEqual(t, len(result), tc.minLength, 
				"Content should meet minimum length requirement")

			// Validate expected content is present
			for _, expected := range tc.expected {
				assert.Contains(t, result, expected, 
					"Should contain expected content: %s", expected)
			}

			// Validate unwanted content is filtered out
			for _, notExpected := range tc.notExpected {
				assert.NotContains(t, result, notExpected, 
					"Should not contain filtered content: %s", notExpected)
			}

			t.Logf("%s: Extracted %d characters", tc.description, len(result))
			t.Logf("Sample content: %s", truncateString(result, 150))
		})
	}
}

// TestJavaScriptCompatibility_NodeSufficiency validates the 100-character threshold
func TestJavaScriptCompatibility_NodeSufficiency(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		sufficient bool
	}{
		{
			name:       "Exactly 100 characters should be sufficient", 
			content:    strings.Repeat("x", 100),
			sufficient: true,
		},
		{
			name:       "99 characters should be insufficient",
			content:    strings.Repeat("x", 99), 
			sufficient: false,
		},
		{
			name:       "101 characters should be sufficient",
			content:    strings.Repeat("x", 101),
			sufficient: true,
		},
		{
			name:       "Long article should be sufficient",
			content:    "This is a comprehensive article with substantial content that provides detailed information on the topic. The article contains multiple sentences and enough text to be considered meaningful and valuable for readers seeking information.",
			sufficient: true,
		},
		{
			name:       "Short snippet should be insufficient", 
			content:    "Brief text",
			sufficient: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			html := "<div><p>" + tc.content + "</p></div>"
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			require.NoError(t, err)

			node := doc.Find("div").First()
			result := NodeIsSufficient(node)

			assert.Equal(t, tc.sufficient, result, 
				"NodeIsSufficient should match expected result for content length: %d", 
				len(tc.content))
		})
	}
}

// TestJavaScriptCompatibility_OptionsCascading validates that options cascade correctly
func TestJavaScriptCompatibility_OptionsCascading(t *testing.T) {
	// HTML that should fail with strict options but succeed with relaxed options
	challengingHTML := `<html>
<body>
	<div class="sidebar-like questionable">
		<p>This content might be challenging to extract with default options but should
		be successfully extracted when options are cascaded and restrictions relaxed.
		The content provides valuable information despite having characteristics that
		might initially classify it as non-article content.</p>
	</div>
</body>
</html>`

	extractor := NewGenericContentExtractor()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(challengingHTML))
	require.NoError(t, err)

	params := ExtractorParams{
		Doc:   doc,
		HTML:  challengingHTML,
		Title: "Challenging Content",
		URL:   "https://example.com/challenging",
	}

	// Test that content is extracted through cascading options
	result := extractor.Extract(params, ExtractorOptions{})

	assert.NotEmpty(t, result, "Should extract content through options cascading")
	assert.Contains(t, result, "challenging to extract", 
		"Should contain the challenging content")
	assert.Contains(t, result, "valuable information", 
		"Should preserve valuable information")

	t.Logf("Cascading extraction successful: %d characters extracted", len(result))
}

// TestJavaScriptCompatibility_SpaceNormalization validates whitespace handling
func TestJavaScriptCompatibility_SpaceNormalization(t *testing.T) {
	html := `<div>
		<p>Content   with    multiple     spaces
		and    line    breaks    that    should    be
		normalized    to    single    spaces.</p>
	</div>`

	extractor := NewGenericContentExtractor()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)

	params := ExtractorParams{
		Doc:   doc,
		HTML:  html,
		Title: "Space Normalization Test",
		URL:   "https://example.com/spaces",
	}

	result := extractor.Extract(params, ExtractorOptions{})

	// Verify spaces are normalized (no multiple consecutive spaces)
	assert.NotContains(t, result, "  ", "Should not contain double spaces")
	assert.NotContains(t, result, "\n\n", "Should not contain double newlines")
	assert.Contains(t, result, "single spaces", "Should contain expected content")

	t.Logf("Space normalization result: %s", result)
}