package parser

import (
	"strings"
	"testing"
	"time"
)

func TestPointerOptimizedMercury_NewPointerOptimized(t *testing.T) {
	// Test with nil options
	parser1 := NewPointerOptimized()
	if parser1 == nil {
		t.Fatal("Expected non-nil parser")
	}

	if !parser1.Mercury.options.FetchAllPages {
		t.Error("Expected FetchAllPages to be true by default")
	}

	if !parser1.Mercury.options.Fallback {
		t.Error("Expected Fallback to be true by default")
	}

	if parser1.Mercury.options.ContentType != "html" {
		t.Errorf("Expected ContentType to be 'html', got '%s'", parser1.Mercury.options.ContentType)
	}

	// Test with custom options
	customOpts := &ParserOptions{
		FetchAllPages: false,
		Fallback:      false,
		ContentType:   "markdown",
		Headers:       map[string]string{"User-Agent": "test"},
	}

	parser2 := NewPointerOptimized(customOpts)
	if parser2.Mercury.options.FetchAllPages {
		t.Error("Expected FetchAllPages to be false")
	}

	if parser2.Mercury.options.Fallback {
		t.Error("Expected Fallback to be false")
	}

	if parser2.Mercury.options.ContentType != "markdown" {
		t.Errorf("Expected ContentType to be 'markdown', got '%s'", parser2.Mercury.options.ContentType)
	}
}

func TestOptimizedExtractorOptions_ToValue(t *testing.T) {
	url := "https://example.com"
	html := "<html><body>test</body></html>"
	metaCache := map[string]string{"title": "Test"}
	fallback := true
	contentType := "html"

	opts := &OptimizedExtractorOptions{
		URL:         &url,
		HTML:        &html,
		MetaCache:   &metaCache,
		Fallback:    &fallback,
		ContentType: &contentType,
	}

	value := opts.ToValue()

	if value.URL != url {
		t.Errorf("Expected URL '%s', got '%s'", url, value.URL)
	}

	if value.HTML != html {
		t.Errorf("Expected HTML '%s', got '%s'", html, value.HTML)
	}

	if len(value.MetaCache) != 1 || value.MetaCache["title"] != "Test" {
		t.Error("MetaCache not converted correctly")
	}

	if !value.Fallback {
		t.Error("Expected Fallback to be true")
	}

	if value.ContentType != contentType {
		t.Errorf("Expected ContentType '%s', got '%s'", contentType, value.ContentType)
	}
}

func TestOptimizedExtractorOptions_ToValueNil(t *testing.T) {
	opts := &OptimizedExtractorOptions{}
	value := opts.ToValue()

	// Test that nil pointers result in zero values
	if value.URL != "" {
		t.Errorf("Expected empty URL, got '%s'", value.URL)
	}

	if value.HTML != "" {
		t.Errorf("Expected empty HTML, got '%s'", value.HTML)
	}

	if value.MetaCache != nil {
		t.Error("Expected nil MetaCache")
	}

	if value.Fallback {
		t.Error("Expected Fallback to be false")
	}

	if value.ContentType != "" {
		t.Errorf("Expected empty ContentType, got '%s'", value.ContentType)
	}
}

func TestBatchOptionsOptimized(t *testing.T) {
	urls := []string{
		"https://example.com/page1",
		"https://example.com/page2",
	}

	opts := &ParserOptions{
		ContentType: "html",
		Fallback:    true,
	}

	concurrency := 5
	timeout := 30

	batchOpts := &BatchOptionsOptimized{
		URLs:        &urls,
		Options:     opts,
		Concurrency: &concurrency,
		Timeout:     &timeout,
	}

	// Test that all fields are set correctly
	if len(*batchOpts.URLs) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(*batchOpts.URLs))
	}

	if batchOpts.Options.ContentType != "html" {
		t.Errorf("Expected ContentType 'html', got '%s'", batchOpts.Options.ContentType)
	}

	if *batchOpts.Concurrency != 5 {
		t.Errorf("Expected Concurrency 5, got %d", *batchOpts.Concurrency)
	}

	if *batchOpts.Timeout != 30 {
		t.Errorf("Expected Timeout 30, got %d", *batchOpts.Timeout)
	}
}

func TestResultOptimizer_MergeResultsOptimized(t *testing.T) {
	optimizer := &ResultOptimizer{}

	// Test merging with nil primary
	result1 := optimizer.MergeResultsOptimized(nil)
	if result1 == nil {
		t.Error("Expected non-nil result")
	}

	// Test merging with primary and additional results
	now := time.Now()
	primary := &Result{
		Title:   "Primary Title",
		URL:     "https://example.com",
		Content: "Primary content",
	}

	additional1 := &Result{
		Author:        "Author Name",
		DatePublished: &now,
		LeadImageURL:  "https://example.com/image.jpg",
	}

	additional2 := &Result{
		Title:       "Additional Title", // Should not override primary
		Dek:         "Description",
		WordCount:   500,
		Extended:    map[string]interface{}{"custom": "value"},
	}

	merged := optimizer.MergeResultsOptimized(primary, additional1, additional2)

	// Verify primary fields are preserved
	if merged.Title != "Primary Title" {
		t.Errorf("Expected primary title preserved, got '%s'", merged.Title)
	}

	if merged.URL != "https://example.com" {
		t.Errorf("Expected primary URL preserved, got '%s'", merged.URL)
	}

	if merged.Content != "Primary content" {
		t.Errorf("Expected primary content preserved, got '%s'", merged.Content)
	}

	// Verify additional fields are merged
	if merged.Author != "Author Name" {
		t.Errorf("Expected author from additional1, got '%s'", merged.Author)
	}

	if merged.DatePublished == nil || !merged.DatePublished.Equal(now) {
		t.Error("Expected DatePublished from additional1")
	}

	if merged.LeadImageURL != "https://example.com/image.jpg" {
		t.Errorf("Expected LeadImageURL from additional1, got '%s'", merged.LeadImageURL)
	}

	if merged.Dek != "Description" {
		t.Errorf("Expected Dek from additional2, got '%s'", merged.Dek)
	}

	if merged.WordCount != 500 {
		t.Errorf("Expected WordCount from additional2, got %d", merged.WordCount)
	}

	// Verify extended fields are merged
	if merged.Extended == nil || merged.Extended["custom"] != "value" {
		t.Error("Expected Extended fields to be merged")
	}
}

func TestResultOptimizer_CloneResultOptimized(t *testing.T) {
	optimizer := &ResultOptimizer{}

	// Test cloning nil result
	cloned := optimizer.CloneResultOptimized(nil)
	if cloned != nil {
		t.Error("Expected nil result for nil input")
	}

	// Test cloning complete result
	now := time.Now()
	original := &Result{
		Title:         "Original Title",
		Content:       "Original content",
		Author:        "Original Author",
		DatePublished: &now,
		LeadImageURL:  "https://example.com/image.jpg",
		Dek:           "Original description",
		URL:           "https://example.com",
		Domain:        "example.com",
		Excerpt:       "Original excerpt",
		WordCount:     250,
		Direction:     "ltr",
		TotalPages:    1,
		RenderedPages: 1,
		Extended:      map[string]interface{}{"custom": "value", "number": 42},
	}

	cloned = optimizer.CloneResultOptimized(original)

	// Verify all fields are copied
	if cloned.Title != original.Title {
		t.Errorf("Expected Title '%s', got '%s'", original.Title, cloned.Title)
	}

	if cloned.Content != original.Content {
		t.Errorf("Expected Content '%s', got '%s'", original.Content, cloned.Content)
	}

	if cloned.Author != original.Author {
		t.Errorf("Expected Author '%s', got '%s'", original.Author, cloned.Author)
	}

	if cloned.DatePublished == nil || !cloned.DatePublished.Equal(*original.DatePublished) {
		t.Error("Expected DatePublished to be cloned correctly")
	}

	// Verify deep copy of DatePublished (different memory addresses)
	if cloned.DatePublished == original.DatePublished {
		t.Error("Expected DatePublished to be deep copied, not same reference")
	}

	// Verify Extended map is deep copied
	if cloned.Extended == nil {
		t.Error("Expected Extended map to be copied")
	}

	if len(cloned.Extended) != len(original.Extended) {
		t.Errorf("Expected Extended map length %d, got %d", len(original.Extended), len(cloned.Extended))
	}

	if cloned.Extended["custom"] != "value" || cloned.Extended["number"] != 42 {
		t.Error("Expected Extended map values to be copied correctly")
	}

	// Verify it's a different map instance
	if &cloned.Extended == &original.Extended {
		t.Error("Expected Extended map to be deep copied, not same reference")
	}

	// Test modification doesn't affect original
	cloned.Title = "Modified Title"
	cloned.Extended["new"] = "added"

	if original.Title == "Modified Title" {
		t.Error("Modifying cloned result should not affect original")
	}

	if _, exists := original.Extended["new"]; exists {
		t.Error("Modifying cloned Extended map should not affect original")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Test NewOptimizedParser convenience function
	opts := &ParserOptions{ContentType: "markdown"}
	parser := NewOptimizedParser(opts)

	if parser == nil {
		t.Fatal("Expected non-nil parser from convenience function")
	}

	if parser.Mercury.options.ContentType != "markdown" {
		t.Errorf("Expected ContentType 'markdown', got '%s'", parser.Mercury.options.ContentType)
	}

	// Test MergeResults convenience function
	primary := &Result{Title: "Primary"}
	additional := &Result{Author: "Author"}

	merged := MergeResults(primary, additional)
	if merged.Title != "Primary" || merged.Author != "Author" {
		t.Error("MergeResults convenience function failed")
	}

	// Test CloneResult convenience function
	cloned := CloneResult(primary)
	if cloned.Title != "Primary" {
		t.Error("CloneResult convenience function failed")
	}

	// Verify it's a different instance
	cloned.Title = "Modified"
	if primary.Title == "Modified" {
		t.Error("CloneResult should create independent copy")
	}
}

// Benchmark tests to verify performance improvements
func BenchmarkMergeResultsOptimized(b *testing.B) {
	optimizer := &ResultOptimizer{}
	
	primary := &Result{
		Title:   "Primary Title",
		Content: "Primary content",
		URL:     "https://example.com",
	}

	additional := &Result{
		Author:       "Author Name",
		LeadImageURL: "https://example.com/image.jpg",
		Extended:     map[string]interface{}{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = optimizer.MergeResultsOptimized(primary, additional)
	}
}

func BenchmarkCloneResultOptimized(b *testing.B) {
	optimizer := &ResultOptimizer{}
	
	result := &Result{
		Title:         "Title",
		Content:       strings.Repeat("content ", 100), // Larger content
		Author:        "Author",
		DatePublished: &time.Time{},
		Extended:      map[string]interface{}{"key1": "value1", "key2": "value2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = optimizer.CloneResultOptimized(result)
	}
}

func BenchmarkOptimizedExtractorOptionsToValue(b *testing.B) {
	url := "https://example.com"
	html := "<html><body>test</body></html>"
	metaCache := map[string]string{"title": "Test"}
	fallback := true
	contentType := "html"

	opts := &OptimizedExtractorOptions{
		URL:         &url,
		HTML:        &html,
		MetaCache:   &metaCache,
		Fallback:    &fallback,
		ContentType: &contentType,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opts.ToValue()
	}
}