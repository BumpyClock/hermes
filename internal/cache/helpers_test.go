package cache

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

const helperTestHTML = `
<html>
<head><title>Test Document</title></head>
<body>
	<div id="main" class="container primary">
		<h1 class="title">Main Title</h1>
		<div class="content">
			<p class="text">First paragraph with <a href="/link1">link</a></p>
			<p class="text">Second paragraph</p>
			<ul class="list">
				<li>Item 1</li>
				<li>Item 2</li>
			</ul>
		</div>
	</div>
</body>
</html>
`

func createHelperTestDocument() *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(helperTestHTML))
	if err != nil {
		panic(err)
	}
	return doc
}

func TestCachedElementOperations_CachedFind(t *testing.T) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache() // Start with clean cache
	
	doc := createHelperTestDocument()
	mainDiv := doc.Find("#main")

	// Test cached find
	paragraphs := ops.CachedFind(mainDiv, "p")
	if paragraphs.Length() != 2 {
		t.Errorf("Expected 2 paragraphs, got %d", paragraphs.Length())
	}

	// Test cache hit - should get same result
	paragraphs2 := ops.CachedFind(mainDiv, "p")
	if paragraphs2.Length() != 2 {
		t.Errorf("Expected 2 paragraphs from cache, got %d", paragraphs2.Length())
	}

	// Verify cache statistics
	stats := ops.GetCacheStats()
	if stats.Hits == 0 {
		t.Error("Expected cache hits")
	}
}

func TestCachedElementOperations_CachedText(t *testing.T) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache()
	
	doc := createHelperTestDocument()
	title := doc.Find("h1.title")

	// Test cached text
	text := ops.CachedText(title)
	if text != "Main Title" {
		t.Errorf("Expected 'Main Title', got '%s'", text)
	}

	// Test cache hit
	text2 := ops.CachedText(title)
	if text2 != "Main Title" {
		t.Errorf("Expected 'Main Title' from cache, got '%s'", text2)
	}

	// Verify cache was used
	stats := ops.GetCacheStats()
	if stats.Hits == 0 {
		t.Error("Expected cache hits for text operations")
	}
}

func TestCachedElementOperations_CachedAttr(t *testing.T) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache()
	
	doc := createHelperTestDocument()
	mainDiv := doc.Find("#main")

	// Test cached attribute
	id, exists := ops.CachedAttr(mainDiv, "id")
	if !exists || id != "main" {
		t.Errorf("Expected id='main', got id='%s', exists=%t", id, exists)
	}

	// Test cache hit
	id2, exists2 := ops.CachedAttr(mainDiv, "id")
	if !exists2 || id2 != "main" {
		t.Errorf("Expected id='main' from cache, got id='%s', exists=%t", id2, exists2)
	}

	// Test non-existent attribute
	nonExistent, exists := ops.CachedAttr(mainDiv, "data-nonexistent")
	if exists || nonExistent != "" {
		t.Errorf("Expected non-existent attribute, got '%s', exists=%t", nonExistent, exists)
	}
}

func TestCachedElementOperations_CachedHasClass(t *testing.T) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache()
	
	doc := createHelperTestDocument()
	mainDiv := doc.Find("#main")

	// Test class checking
	hasContainer := ops.CachedHasClass(mainDiv, "container")
	if !hasContainer {
		t.Error("Expected element to have 'container' class")
	}

	hasPrimary := ops.CachedHasClass(mainDiv, "primary")
	if !hasPrimary {
		t.Error("Expected element to have 'primary' class")
	}

	hasNonExistent := ops.CachedHasClass(mainDiv, "nonexistent")
	if hasNonExistent {
		t.Error("Expected element not to have 'nonexistent' class")
	}
}

func TestCachedElementOperations_BatchCachedFind(t *testing.T) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache()
	
	doc := createHelperTestDocument()
	mainDiv := doc.Find("#main")

	selectors := []string{"p", "h1", "ul", "li"}
	results := ops.BatchCachedFind(mainDiv, selectors)

	if len(results) != len(selectors) {
		t.Errorf("Expected %d results, got %d", len(selectors), len(results))
	}

	if results["p"].Length() != 2 {
		t.Errorf("Expected 2 paragraphs, got %d", results["p"].Length())
	}

	if results["h1"].Length() != 1 {
		t.Errorf("Expected 1 h1, got %d", results["h1"].Length())
	}

	if results["li"].Length() != 2 {
		t.Errorf("Expected 2 list items, got %d", results["li"].Length())
	}

	// Test batch query again - should hit cache
	results2 := ops.BatchCachedFind(mainDiv, selectors)
	if results2["p"].Length() != 2 {
		t.Errorf("Expected 2 paragraphs from cache, got %d", results2["p"].Length())
	}
}

func TestCachedElementOperations_OptimizedLinkDensity(t *testing.T) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache()
	
	doc := createHelperTestDocument()
	
	// Test with element containing links
	content := doc.Find(".content")
	density := ops.OptimizedLinkDensity(content)
	
	if density <= 0 {
		t.Error("Expected link density to be greater than 0")
	}

	// Test with element without links
	title := doc.Find("h1")
	titleDensity := ops.OptimizedLinkDensity(title)
	
	if titleDensity != 0 {
		t.Errorf("Expected title link density to be 0, got %f", titleDensity)
	}

	// Test cache effectiveness by running again
	density2 := ops.OptimizedLinkDensity(content)
	if density2 != density {
		t.Errorf("Expected same density from cache, got %f vs %f", density, density2)
	}
}

func TestGlobalCachedFunctions(t *testing.T) {
	// Clear global cache
	GlobalCachedOps.ClearElementCache()
	
	doc := createHelperTestDocument()
	mainDiv := doc.Find("#main")

	// Test global functions
	paragraphs := CachedFind(mainDiv, "p")
	if paragraphs.Length() != 2 {
		t.Errorf("Expected 2 paragraphs, got %d", paragraphs.Length())
	}

	title := doc.Find("h1")
	text := CachedText(title)
	if text != "Main Title" {
		t.Errorf("Expected 'Main Title', got '%s'", text)
	}

	id, exists := CachedAttr(mainDiv, "id")
	if !exists || id != "main" {
		t.Errorf("Expected id='main', got id='%s', exists=%t", id, exists)
	}

	hasClass := CachedHasClass(mainDiv, "container")
	if !hasClass {
		t.Error("Expected element to have 'container' class")
	}

	density := OptimizedLinkDensity(doc.Find(".content"))
	if density <= 0 {
		t.Error("Expected link density to be greater than 0")
	}

	selectors := []string{"p", "h1"}
	batchResults := BatchCachedFind(mainDiv, selectors)
	if len(batchResults) != 2 {
		t.Errorf("Expected 2 batch results, got %d", len(batchResults))
	}
}

func TestCachedElementOperations_EmptySelection(t *testing.T) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache()
	
	doc := createHelperTestDocument()
	empty := doc.Find(".nonexistent")

	// Test operations on empty selection
	result := ops.CachedFind(empty, "p")
	if result.Length() != 0 {
		t.Errorf("Expected empty result, got %d elements", result.Length())
	}

	text := ops.CachedText(empty)
	if text != "" {
		t.Errorf("Expected empty text, got '%s'", text)
	}

	_, exists := ops.CachedAttr(empty, "id")
	if exists {
		t.Error("Expected no attributes for empty selection")
	}

	hasClass := ops.CachedHasClass(empty, "test")
	if hasClass {
		t.Error("Expected no classes for empty selection")
	}
}

// Benchmark tests to verify caching performance benefits
func BenchmarkCachedFind(b *testing.B) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache()
	
	doc := createHelperTestDocument()
	mainDiv := doc.Find("#main")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ops.CachedFind(mainDiv, "p")
	}
}

func BenchmarkUncachedFind(b *testing.B) {
	doc := createHelperTestDocument()
	mainDiv := doc.Find("#main")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mainDiv.Find("p")
	}
}

func BenchmarkCachedText(b *testing.B) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache()
	
	doc := createHelperTestDocument()
	title := doc.Find("h1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ops.CachedText(title)
	}
}

func BenchmarkUncachedText(b *testing.B) {
	doc := createHelperTestDocument()
	title := doc.Find("h1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		title.Text()
	}
}

func BenchmarkBatchCachedFind(b *testing.B) {
	ops := NewCachedElementOperations()
	ops.ClearElementCache()
	
	doc := createHelperTestDocument()
	mainDiv := doc.Find("#main")
	selectors := []string{"p", "h1", "ul", "li", "a"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ops.BatchCachedFind(mainDiv, selectors)
	}
}

func BenchmarkSequentialUncachedFind(b *testing.B) {
	doc := createHelperTestDocument()
	mainDiv := doc.Find("#main")
	selectors := []string{"p", "h1", "ul", "li", "a"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, selector := range selectors {
			mainDiv.Find(selector)
		}
	}
}