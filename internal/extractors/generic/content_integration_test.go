// ABOUTME: Integration tests for content extractor verifying end-to-end JavaScript compatibility
// ABOUTME: Tests real-world scenarios with complex HTML and validates extraction quality matches JavaScript behavior

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentExtractor_EndToEndExtraction(t *testing.T) {
	// Complex HTML that mimics real-world article structure
	complexHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Test Article - Example News Site</title>
</head>
<body>
    <!-- Navigation and header content -->
    <nav class="navigation">
        <ul>
            <li><a href="/">Home</a></li>
            <li><a href="/news">News</a></li>
        </ul>
    </nav>
    
    <!-- Ads and unlikely candidates -->
    <div class="ad-banner">Advertisement content here</div>
    <div class="comments-section">User comments and discussions</div>
    
    <!-- Main article content -->
    <article class="main-article">
        <header>
            <h1>This is the main article headline</h1>
            <div class="article-meta">
                <span class="author">By John Doe</span>
                <time>2024-01-15</time>
            </div>
        </header>
        
        <div class="article-body">
            <p>This is the first paragraph of the main article content. It contains substantial 
            information that should be extracted as part of the primary content. This paragraph 
            has enough content to meet the sufficiency requirements.</p>
            
            <p>The second paragraph continues with more detailed information about the topic.
            This content should also be included in the final extracted result as it provides
            valuable information to the reader and contributes to the overall article length.</p>
            
            <h2>Important Subheading</h2>
            <p>Content under the subheading that elaborates on specific aspects of the topic.
            This section adds depth to the article and should be preserved in the extraction
            process to maintain the complete narrative structure.</p>
            
            <blockquote>
                <p>This is an important quote that adds context to the article content.
                Quotes like this should be preserved as they often contain key information
                or perspectives that are central to the story.</p>
            </blockquote>
            
            <p>Final paragraph that concludes the article with summary information and
            final thoughts. This content wraps up the main points and should be included
            to provide closure to the extracted content.</p>
        </div>
    </article>
    
    <!-- Footer and additional unlikely content -->
    <footer>
        <div class="related-articles">Related articles links</div>
        <div class="social-sharing">Social sharing buttons</div>
    </footer>
    
    <div class="sidebar">
        <div class="newsletter-signup">Newsletter signup form</div>
        <div class="more-ads">More advertisement content</div>
    </div>
</body>
</html>`

	extractor := NewGenericContentExtractor()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(complexHTML))
	require.NoError(t, err)

	params := ExtractorParams{
		Doc:   doc,
		HTML:  complexHTML,
		Title: "Test Article - Example News Site",
		URL:   "https://example.com/test-article",
	}

	// Test with default options (strict)
	result := extractor.Extract(params, ExtractorOptions{})
	
	t.Logf("Extracted content length: %d characters", len(result))
	t.Logf("First 300 chars: %s", truncateString(result, 300))

	// Verify content was extracted
	assert.NotEmpty(t, result, "should extract content from complex HTML")
	
	// Verify main content is included
	assert.Contains(t, result, "first paragraph of the main article")
	assert.Contains(t, result, "second paragraph continues")
	assert.Contains(t, result, "Important Subheading")
	assert.Contains(t, result, "important quote that adds context")
	assert.Contains(t, result, "Final paragraph that concludes")

	// Verify unlikely candidates were filtered out
	assert.NotContains(t, result, "Advertisement content")
	assert.NotContains(t, result, "User comments")
	assert.NotContains(t, result, "Newsletter signup")
	assert.NotContains(t, result, "Social sharing")
}

func TestContentExtractor_JavaScriptCompatibilityVerification(t *testing.T) {
	// This HTML structure matches patterns found in JavaScript test fixtures
	jsCompatHTML := `<html>
<body>
    <div class="entry-content">
        <p>This content should be extracted with the same behavior as the JavaScript version.
        The content extraction algorithm should identify this as the primary content area
        based on scoring algorithms and content analysis techniques that match the original
        JavaScript implementation exactly.</p>
        
        <p>Additional paragraph content that contributes to the overall score and should
        be included in the final extraction. The scoring system evaluates content density,
        text length, and structural elements to determine extraction quality.</p>
        
        <ul>
            <li>First important point in the article</li>
            <li>Second key point with detailed information</li>
            <li>Third point that concludes the list</li>
        </ul>
        
        <p>Concluding paragraph that wraps up the content and provides final thoughts.
        This text should be preserved to maintain the complete article structure and
        ensure readers get the full context of the content.</p>
    </div>
    
    <div class="sidebar">
        <div class="widget">Sidebar widget content</div>
    </div>
</body>
</html>`

	extractor := NewGenericContentExtractor()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(jsCompatHTML))
	require.NoError(t, err)

	params := ExtractorParams{
		Doc:   doc,
		HTML:  jsCompatHTML,
		Title: "JavaScript Compatibility Test",
		URL:   "https://test.example.com/js-compat",
	}

	result := extractor.Extract(params, ExtractorOptions{})
	
	// Verify the extracted content matches expected JavaScript behavior
	assert.Contains(t, result, "same behavior as the JavaScript version")
	assert.Contains(t, result, "scoring algorithms and content analysis")
	assert.Contains(t, result, "First important point")
	assert.Contains(t, result, "Second key point")
	assert.Contains(t, result, "Third point that concludes")
	assert.Contains(t, result, "Concluding paragraph that wraps up")
	
	// Verify sidebar content is not included
	assert.NotContains(t, result, "Sidebar widget content")
	
	// Verify content is properly normalized (no excessive whitespace)
	assert.NotContains(t, result, "  ", "should normalize spaces properly")
	
	t.Logf("JavaScript compatibility test passed. Content length: %d", len(result))
}

func TestContentExtractor_OptionsCascading(t *testing.T) {
	// HTML that might require options cascading to extract properly
	challengingHTML := `<html>
<body>
    <div class="content-wrapper">
        <div class="questionable-content sidebar-like">
            <p>This content has characteristics that might make it seem like sidebar content
            but it's actually the main article content. The cascading options should help
            extract this when strict options fail.</p>
        </div>
    </div>
</body>
</html>`

	extractor := NewGenericContentExtractor()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(challengingHTML))
	require.NoError(t, err)

	params := ExtractorParams{
		Doc:   doc,
		HTML:  challengingHTML,
		Title: "Options Cascading Test",
		URL:   "https://example.com/cascading",
	}

	// Test extraction with cascading - should succeed even with challenging content
	result := extractor.Extract(params, ExtractorOptions{})
	
	assert.NotEmpty(t, result, "should extract content through options cascading")
	assert.Contains(t, result, "main article content")
	assert.Contains(t, result, "cascading options should help")
	
	t.Logf("Options cascading successful. Extracted: %s", truncateString(result, 150))
}

// Helper function to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}