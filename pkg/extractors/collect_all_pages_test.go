// ABOUTME: Comprehensive test suite for multi-page article collection system
// ABOUTME: Tests pagination logic, URL deduplication, content merging, and safety limits with JavaScript compatibility

package extractors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test HTML snippets for multi-page collection scenarios
const (
	// First page HTML with next page URL
	firstPageHTML = `
		<html>
			<head><title>Multi-Page Article</title></head>
			<body>
				<div class="article-content">
					<p>This is the first page of the multi-page article.</p>
					<p>It contains some important initial content.</p>
				</div>
				<a href="/page-2" rel="next">Next Page</a>
			</body>
		</html>`

	// Second page HTML with next page URL
	secondPageHTML = `
		<html>
			<head><title>Multi-Page Article</title></head>
			<body>
				<div class="article-content">
					<p>This is the second page of the multi-page article.</p>
					<p>It continues the story from the first page.</p>
				</div>
				<a href="/page-3" rel="next">Next Page</a>
			</body>
		</html>`

	// Third page HTML without next page URL (final page)
	thirdPageHTML = `
		<html>
			<head><title>Multi-Page Article</title></head>
			<body>
				<div class="article-content">
					<p>This is the final page of the multi-page article.</p>
					<p>It concludes the story.</p>
				</div>
			</body>
		</html>`

	// Single page HTML without next page URL
	singlePageHTML = `
		<html>
			<head><title>Single Page Article</title></head>
			<body>
				<div class="article-content">
					<p>This is a single page article with no pagination.</p>
				</div>
			</body>
		</html>`

	// Page with circular reference (for cycle detection testing)
	circularPageHTML = `
		<html>
			<head><title>Circular Page</title></head>
			<body>
				<div class="article-content">
					<p>This page links back to a previous page.</p>
				</div>
				<a href="/page-1" rel="next">Back to First Page</a>
			</body>
		</html>`
)

// MockResource provides a mock implementation of the Resource interface
type MockResource struct {
	PageResponses map[string]string
	CallCount     int
	CallLog       []string
}

// Create simulates fetching pages based on URL patterns
func (m *MockResource) Create(url string, preparedResponse string, parsedURL interface{}, headers map[string]string) (*goquery.Document, error) {
	m.CallCount++
	m.CallLog = append(m.CallLog, url)

	// Return pre-configured HTML based on URL
	if html, exists := m.PageResponses[url]; exists {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		return doc, err
	}

	return nil, fmt.Errorf("page not found: %s", url)
}

// MockExtractor provides a mock implementation that returns predictable results
type MockExtractor struct {
	ExtractorConfig map[string]interface{}
	PageContent     map[string]string
	NextPageURLs    map[string]string
}

func (m *MockExtractor) Extract(extractor interface{}, opts ExtractOptions) interface{} {
	url := opts.URL

	// Return mock extraction results based on URL
	content := "Default content"
	if c, exists := m.PageContent[url]; exists {
		content = c
	}

	nextPageURL := ""
	if next, exists := m.NextPageURLs[url]; exists {
		nextPageURL = next
	}

	return map[string]interface{}{
		"title":         "Test Title",
		"content":       content,
		"author":        "Test Author",
		"next_page_url": nextPageURL,
		"url":           url,
		"domain":        "example.com",
	}
}

func TestCollectAllPages_SinglePage(t *testing.T) {
	t.Run("no next_page_url should return original result", func(t *testing.T) {
		// Setup mock dependencies
		mockResource := &MockResource{
			PageResponses: map[string]string{},
		}
		
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(singlePageHTML))
		require.NoError(t, err)

		originalResult := map[string]interface{}{
			"title":         "Single Page Article",
			"content":       "<p>Original content</p>",
			"author":        "Test Author",
			"next_page_url": nil, // No next page
		}

		// Call CollectAllPages
		result := CollectAllPages(CollectAllPagesOptions{
			NextPageURL: "",
			HTML:        singlePageHTML,
			Doc:         doc,
			MetaCache:   map[string]interface{}{},
			Result:      originalResult,
			Extractor:   map[string]interface{}{"domain": "example.com"},
			Title:       "Single Page Article",
			URL:         "http://example.com/single-page",
			Resource:    mockResource,
		})

		// Verify result
		assert.Equal(t, 1, result["total_pages"])
		assert.Equal(t, 1, result["rendered_pages"])
		assert.Equal(t, originalResult["content"], result["content"])
		assert.NotNil(t, result["word_count"])
		
		// Verify no additional resource calls were made
		assert.Equal(t, 0, mockResource.CallCount)
	})
}

func TestCollectAllPages_MultiplePages(t *testing.T) {
	t.Run("should collect and merge multiple pages with separators", func(t *testing.T) {
		// Setup mock dependencies
		mockResource := &MockResource{
			PageResponses: map[string]string{
				"http://example.com/page-2": secondPageHTML,
				"http://example.com/page-3": thirdPageHTML,
			},
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(firstPageHTML))
		require.NoError(t, err)

		originalResult := map[string]interface{}{
			"title":         "Multi-Page Article",
			"content":       "<p>First page content</p>",
			"next_page_url": "http://example.com/page-2",
		}

		// Call CollectAllPages
		result := CollectAllPages(CollectAllPagesOptions{
			NextPageURL: "http://example.com/page-2",
			HTML:        firstPageHTML,
			Doc:         doc,
			MetaCache:   map[string]interface{}{},
			Result:      originalResult,
			Extractor:   map[string]interface{}{"domain": "example.com"},
			Title:       "Multi-Page Article",
			URL:         "http://example.com/page-1",
			Resource:    mockResource,
			RootExtractor: &RootExtractorInterface{},
		})

		// Verify result structure
		assert.Equal(t, 3, result["total_pages"])
		assert.Equal(t, 3, result["rendered_pages"])
		assert.NotNil(t, result["word_count"])

		// Verify content merging with separators
		content := result["content"].(string)
		assert.Contains(t, content, "<p>First page content</p>")
		assert.Contains(t, content, "<hr><h4>Page 2</h4>")
		assert.Contains(t, content, "<p>Second page content</p>")
		assert.Contains(t, content, "<hr><h4>Page 3</h4>")
		assert.Contains(t, content, "<p>Third page content</p>")

		// Verify resource calls
		assert.Equal(t, 2, mockResource.CallCount)
		assert.Contains(t, mockResource.CallLog, "http://example.com/page-2")
		assert.Contains(t, mockResource.CallLog, "http://example.com/page-3")
	})
}

func TestCollectAllPages_SafetyLimit(t *testing.T) {
	t.Run("should stop at 26 pages to prevent infinite loops", func(t *testing.T) {
		// Setup mock resource that always returns a next page
		mockResource := &MockResource{
			PageResponses: make(map[string]string),
		}

		// Generate 30 pages worth of responses (more than the 26 limit)
		for i := 2; i <= 30; i++ {
			url := fmt.Sprintf("http://example.com/page-%d", i)
			nextURL := fmt.Sprintf("http://example.com/page-%d", i+1)
			
			mockResource.PageResponses[url] = fmt.Sprintf(`
				<html><body>
					<div class="content"><p>Page %d content</p></div>
					<a href="%s" rel="next">Next</a>
				</body></html>`, i, nextURL)
			
			// Mock extractor setup removed for simplicity
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(firstPageHTML))
		require.NoError(t, err)

		originalResult := map[string]interface{}{
			"title":         "Infinite Article",
			"content":       "<p>First page content</p>",
			"next_page_url": "http://example.com/page-2",
		}

		// Call CollectAllPages
		result := CollectAllPages(CollectAllPagesOptions{
			NextPageURL: "http://example.com/page-2",
			HTML:        firstPageHTML,
			Doc:         doc,
			MetaCache:   map[string]interface{}{},
			Result:      originalResult,
			Extractor:   map[string]interface{}{"domain": "example.com"},
			Title:       "Infinite Article",
			URL:         "http://example.com/page-1",
			Resource:    mockResource,
			RootExtractor: &RootExtractorInterface{},
		})

		// Verify safety limit was enforced
		assert.Equal(t, 26, result["total_pages"])
		assert.Equal(t, 26, result["rendered_pages"])
		
		// Should have made exactly 25 resource calls (pages 2-26)
		assert.Equal(t, 25, mockResource.CallCount)
		
		// Content should contain exactly 26 pages worth of content
		content := result["content"].(string)
		assert.Contains(t, content, "<p>First page content</p>") // Page 1
		assert.Contains(t, content, "<hr><h4>Page 26</h4>")      // Page 26
		assert.NotContains(t, content, "<hr><h4>Page 27</h4>")   // Should not have page 27
	})
}

func TestCollectAllPages_URLDeduplication(t *testing.T) {
	t.Run("should prevent cycles by tracking previous URLs", func(t *testing.T) {
		mockResource := &MockResource{
			PageResponses: map[string]string{
				"http://example.com/page-2": secondPageHTML,
				"http://example.com/page-1": circularPageHTML, // Links back to page 1
			},
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(firstPageHTML))
		require.NoError(t, err)

		originalResult := map[string]interface{}{
			"title":         "Circular Article",
			"content":       "<p>First page content</p>",
			"next_page_url": "http://example.com/page-2",
		}

		// Call CollectAllPages
		result := CollectAllPages(CollectAllPagesOptions{
			NextPageURL: "http://example.com/page-2",
			HTML:        firstPageHTML,
			Doc:         doc,
			MetaCache:   map[string]interface{}{},
			Result:      originalResult,
			Extractor:   map[string]interface{}{"domain": "example.com"},
			Title:       "Circular Article",
			URL:         "http://example.com/page-1",
			Resource:    mockResource,
			RootExtractor: &RootExtractorInterface{},
		})

		// Should have collected 2 pages (original + page-2), then stopped due to cycle detection
		assert.Equal(t, 2, result["total_pages"])
		assert.Equal(t, 2, result["rendered_pages"])
		
		// Should have made only 1 resource call (for page-2)
		assert.Equal(t, 1, mockResource.CallCount)
		assert.Equal(t, []string{"http://example.com/page-2"}, mockResource.CallLog)

		// Content should contain first two pages only
		content := result["content"].(string)
		assert.Contains(t, content, "<p>First page content</p>")
		assert.Contains(t, content, "<hr><h4>Page 2</h4>")
		assert.Contains(t, content, "<p>Second page content</p>")
		assert.NotContains(t, content, "<hr><h4>Page 3</h4>") // Should not have a third page
	})
}

func TestCollectAllPages_WordCountCalculation(t *testing.T) {
	t.Run("should calculate accurate word count for merged content", func(t *testing.T) {
		mockResource := &MockResource{
			PageResponses: map[string]string{
				"http://example.com/page-2": secondPageHTML,
			},
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(firstPageHTML))
		require.NoError(t, err)

		originalResult := map[string]interface{}{
			"title":         "Word Count Article",
			"content":       "<p>First page has some words</p>",
			"next_page_url": "http://example.com/page-2",
		}

		// Call CollectAllPages
		result := CollectAllPages(CollectAllPagesOptions{
			NextPageURL: "http://example.com/page-2",
			HTML:        firstPageHTML,
			Doc:         doc,
			MetaCache:   map[string]interface{}{},
			Result:      originalResult,
			Extractor:   map[string]interface{}{"domain": "example.com"},
			Title:       "Word Count Article",
			URL:         "http://example.com/page-1",
			Resource:    mockResource,
			RootExtractor: &RootExtractorInterface{},
		})

		// Verify word count is calculated
		wordCount := result["word_count"]
		assert.NotNil(t, wordCount)
		assert.IsType(t, 0, wordCount) // Should be an integer
		assert.Greater(t, wordCount.(int), 0) // Should be positive
		
		// Word count should account for merged content
		// The merged content includes both pages plus separator text
		assert.Greater(t, wordCount.(int), 5) // Should be more than just a few words
	})
}

func TestCollectAllPages_JavaScriptCompatibility(t *testing.T) {
	t.Run("should exactly match JavaScript behavior", func(t *testing.T) {
		// This test verifies that our Go implementation matches the JavaScript version exactly
		
		// JavaScript implementation starts pages counter at 1 (first page already fetched)
		// and increments for each subsequent page
		mockResource := &MockResource{
			PageResponses: map[string]string{
				"http://example.com/page-2": secondPageHTML,
			},
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(firstPageHTML))
		require.NoError(t, err)

		originalResult := map[string]interface{}{
			"title":         "JS Compat Article",
			"content":       "<p>First page content</p>",
			"next_page_url": "http://example.com/page-2",
		}

		// Call CollectAllPages
		result := CollectAllPages(CollectAllPagesOptions{
			NextPageURL: "http://example.com/page-2",
			HTML:        firstPageHTML,
			Doc:         doc,
			MetaCache:   map[string]interface{}{},
			Result:      originalResult,
			Extractor:   map[string]interface{}{"domain": "example.com"},
			Title:       "JS Compat Article",
			URL:         "http://example.com/page-1",
			Resource:    mockResource,
			RootExtractor: &RootExtractorInterface{},
		})

		// JavaScript behavior verification:
		// 1. Pages counter starts at 1 and increments (should be 2 for this test)
		assert.Equal(t, 2, result["total_pages"])
		assert.Equal(t, 2, result["rendered_pages"])

		// 2. Content merging format: `${result.content}<hr><h4>Page ${pages}</h4>${nextPageResult.content}`
		content := result["content"].(string)
		expectedContent := "<p>First page content</p><hr><h4>Page 2</h4><p>Second page content</p>"
		assert.Equal(t, expectedContent, content)

		// 3. Word count calculation using GenericExtractor.word_count with <div> wrapper
		wordCount := result["word_count"]
		assert.NotNil(t, wordCount)
		
		// 4. Result structure should match JavaScript exactly
		requiredFields := []string{"total_pages", "rendered_pages", "word_count"}
		for _, field := range requiredFields {
			assert.Contains(t, result, field, "Field %s should be present", field)
		}
	})
}

func TestCollectAllPages_ErrorHandling(t *testing.T) {
	t.Run("should handle resource fetch failures gracefully", func(t *testing.T) {
		// Mock resource that fails to fetch pages
		mockResource := &MockResource{
			PageResponses: map[string]string{
				// No responses configured - all fetches will fail
			},
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(firstPageHTML))
		require.NoError(t, err)

		originalResult := map[string]interface{}{
			"title":         "Error Handling Article",
			"content":       "<p>First page content</p>",
			"next_page_url": "http://example.com/page-2",
		}

		// Call CollectAllPages with failing resource
		result := CollectAllPages(CollectAllPagesOptions{
			NextPageURL: "http://example.com/page-2",
			HTML:        firstPageHTML,
			Doc:         doc,
			MetaCache:   map[string]interface{}{},
			Result:      originalResult,
			Extractor:   map[string]interface{}{"domain": "example.com"},
			Title:       "Error Handling Article",
			URL:         "http://example.com/page-1",
			Resource:    mockResource,
		})

		// Should return original result when fetch fails
		assert.Equal(t, 1, result["total_pages"])
		assert.Equal(t, 1, result["rendered_pages"])
		assert.Equal(t, originalResult["content"], result["content"])
		
		// Resource should have been called but failed
		assert.Equal(t, 1, mockResource.CallCount)
	})
}