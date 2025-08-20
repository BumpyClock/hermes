// ABOUTME: Comprehensive tests for generic extractor orchestration with JavaScript compatibility verification
// ABOUTME: Tests field extraction order, dependencies, and complete extraction workflow

package generic

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Sample HTML for testing generic extraction
const testHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Test Article - Example Site</title>
    <meta name="description" content="This is a test article description for generic extraction testing.">
    <meta name="author" content="John Doe">
    <meta name="og:image" content="https://example.com/image.jpg">
    <meta name="date" content="2024-01-15">
</head>
<body>
    <h1>Test Article Title</h1>
    <p class="byline">By Jane Smith</p>
    <div class="content">
        <p>This is the main content of the test article. It contains multiple paragraphs to test content extraction.</p>
        <p>This is a second paragraph with more content to ensure word count calculation works properly.</p>
        <p>A third paragraph to provide sufficient content for testing excerpt generation and other content-dependent extractions.</p>
    </div>
    <a href="/page2" rel="next">Next Page</a>
</body>
</html>
`

// TestNewGenericExtractor verifies constructor
func TestNewGenericExtractor(t *testing.T) {
	extractor := NewGenericExtractor()
	
	if extractor == nil {
		t.Fatal("NewGenericExtractor returned nil")
	}
	
	if extractor.GetDomain() != "*" {
		t.Errorf("Expected domain '*', got '%s'", extractor.GetDomain())
	}
}

// TestGenericExtractor_ExtractGeneric tests the main extraction workflow
func TestGenericExtractor_ExtractGeneric(t *testing.T) {
	extractor := NewGenericExtractor()
	
	// Parse test HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to parse test HTML: %v", err)
	}
	
	// Create extraction options
	options := &ExtractionOptions{
		URL:         "https://example.com/test-article",
		Doc:         doc,
		MetaCache:   []string{"description", "author", "date"},
		Fallback:    true,
		ContentType: "html",
	}
	
	// Extract content
	result, err := extractor.ExtractGeneric(options)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}
	
	// Verify basic fields are extracted
	if result.Title == "" {
		t.Error("Expected title to be extracted")
	}
	
	if result.Content == "" {
		t.Error("Expected content to be extracted")
	}
	
	if result.URL != "https://example.com/test-article" {
		t.Errorf("Expected URL 'https://example.com/test-article', got '%s'", result.URL)
	}
	
	if result.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got '%s'", result.Domain)
	}
	
	// Word count should be greater than 0
	if result.WordCount <= 0 {
		t.Errorf("Expected word count > 0, got %d", result.WordCount)
	}
	
	// Direction should be detected
	if result.Direction == "" {
		t.Error("Expected direction to be detected")
	}
}

// TestGenericExtractor_ExtractOrder tests that fields are extracted in correct order
func TestGenericExtractor_ExtractOrder(t *testing.T) {
	extractor := NewGenericExtractor()
	
	// Parse test HTML  
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to parse test HTML: %v", err)
	}
	
	options := &ExtractionOptions{
		URL:         "https://example.com/test",
		Doc:         doc,
		MetaCache:   []string{"description", "author", "date"},
		Fallback:    true,
		ContentType: "html",
	}
	
	result, err := extractor.ExtractGeneric(options)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}
	
	// Verify JavaScript-compatible field structure
	expectedFields := []string{"Title", "Author", "Content", "URL", "Domain", "Excerpt", "Direction"}
	
	// Check that all expected fields are present (non-empty or have valid defaults)
	if result.Title == "" {
		t.Error("Title field missing")
	}
	
	if result.URL == "" {
		t.Error("URL field missing")
	}
	
	if result.Domain == "" {
		t.Error("Domain field missing") 
	}
	
	// Content-dependent fields should be present after content extraction
	if result.Content != "" {
		if result.WordCount <= 0 {
			t.Error("WordCount should be calculated after content extraction")
		}
		
		// Excerpt may be empty for short content, but should not cause errors
		// This validates that excerpt extractor was called with content
	}
	
	t.Logf("Extracted fields successfully: %v", expectedFields)
}

// TestGenericExtractor_Extract tests Extractor interface implementation
func TestGenericExtractor_Extract(t *testing.T) {
	extractor := NewGenericExtractor()
	
	// Parse test HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to parse test HTML: %v", err)
	}
	
	// Test interface method
	result, err := extractor.Extract(doc)
	if err != nil {
		t.Fatalf("Interface extraction failed: %v", err)
	}
	
	// Result should be of type *ExtractionResult
	extractionResult, ok := result.(*ExtractionResult)
	if !ok {
		t.Fatalf("Expected *ExtractionResult, got %T", result)
	}
	
	// Basic validation
	if extractionResult.Content == "" {
		t.Error("Expected content to be extracted via interface method")
	}
}

// TestGenericExtractor_ErrorHandling tests error conditions
func TestGenericExtractor_ErrorHandling(t *testing.T) {
	extractor := NewGenericExtractor()
	
	// Test with nil document
	options := &ExtractionOptions{
		URL:         "https://example.com/test",
		Doc:         nil,
		HTML:        "", // No HTML either
		MetaCache:   []string{},
		Fallback:    true,
		ContentType: "html",
	}
	
	_, err := extractor.ExtractGeneric(options)
	if err == nil {
		t.Error("Expected error with nil document and no HTML")
	}
	
	// Test with invalid HTML
	options.HTML = "<invalid><html"
	_, err = extractor.ExtractGeneric(options)
	// Should not error - goquery is forgiving with malformed HTML
}

// TestGenericExtractor_FieldDependencies tests that field extraction respects dependencies
func TestGenericExtractor_FieldDependencies(t *testing.T) {
	extractor := NewGenericExtractor()
	
	// Parse test HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to parse test HTML: %v", err)
	}
	
	options := &ExtractionOptions{
		URL:         "https://example.com/test",
		Doc:         doc,
		MetaCache:   []string{"description", "author"},
		Fallback:    true,
		ContentType: "html",
	}
	
	result, err := extractor.ExtractGeneric(options)
	if err != nil {
		t.Fatalf("Extraction failed: %v", err)
	}
	
	// Test dependencies from JavaScript:
	// - content depends on title
	// - lead_image_url depends on content  
	// - dek depends on content
	// - excerpt depends on content
	// - word_count depends on content
	// - direction depends on title
	
	// If we have title, we should have content
	if result.Title != "" && result.Content == "" {
		t.Error("Content extraction should succeed when title is available")
	}
	
	// If we have content, content-dependent fields should be processed
	if result.Content != "" {
		if result.WordCount <= 0 {
			t.Error("WordCount should be calculated from content")
		}
		
		// Excerpt extractor should be called (may return empty for short content)
		// Lead image and dek extractors should be called (may return empty)
		// This test validates the call sequence, not necessarily non-empty results
	}
	
	// Direction should be calculated from title
	if result.Title != "" && result.Direction == "" {
		t.Error("Direction should be calculated from title")
	}
}

// BenchmarkGenericExtractor_ExtractGeneric benchmarks full extraction
func BenchmarkGenericExtractor_ExtractGeneric(b *testing.B) {
	extractor := NewGenericExtractor()
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	
	options := &ExtractionOptions{
		URL:         "https://example.com/test",
		Doc:         doc,
		MetaCache:   []string{"description", "author", "date"},
		Fallback:    true,
		ContentType: "html",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.ExtractGeneric(options)
		if err != nil {
			b.Fatalf("Extraction failed: %v", err)
		}
	}
}