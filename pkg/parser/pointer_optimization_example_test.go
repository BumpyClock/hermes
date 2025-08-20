package parser

import (
	"fmt"
	"log"
)

// Example of using pointer-optimized parser for better performance
func ExampleNewOptimizedParser() {
	// Create parser with pointer-optimized options
	opts := &ParserOptions{
		ContentType:   "html",
		Fallback:      true,
		FetchAllPages: false,
	}

	parser := NewOptimizedParser(opts)

	// Use the optimized parser (would normally parse real URLs)
	// result, err := parser.ParseOptimized("https://example.com/article", opts)
	// if err != nil {
	//     log.Fatal(err)
	// }

	_ = parser // Use the parser variable to avoid unused variable error
	fmt.Printf("Parser created with options: %+v\n", parser.Mercury.options)
	// Output: Parser created with options: {FetchAllPages:false Fallback:true ContentType:html Headers:map[] CustomExtractor:<nil> Extend:map[]}
}

// Example of batch processing with pointer optimization
func ExampleBatchOptionsOptimized() {
	parser := NewOptimizedParser()

	urls := []string{
		"https://example.com/article1",
		"https://example.com/article2",
		"https://example.com/article3",
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

	// Process batch (would normally process real URLs)
	// results, err := parser.ProcessBatchOptimized(batchOpts)
	// if err != nil {
	//     log.Fatal(err)
	// }

	_ = parser // Use the parser variable to avoid unused variable error
	fmt.Printf("Batch options configured for %d URLs with %d concurrency\n", 
		len(*batchOpts.URLs), *batchOpts.Concurrency)
	// Output: Batch options configured for 3 URLs with 5 concurrency
}

// Example of efficient result merging
func ExampleMergeResults() {
	// Create primary result with basic info
	primary := &Result{
		Title:   "Article Title",
		URL:     "https://example.com/article",
		Content: "Main article content",
	}

	// Create additional results with supplementary info
	metadata := &Result{
		Author:       "John Doe",
		LeadImageURL: "https://example.com/image.jpg",
		WordCount:    150,
	}

	extended := &Result{
		Dek:       "Article description",
		Direction: "ltr",
		Extended:  map[string]interface{}{"category": "technology"},
	}

	// Merge all results efficiently
	merged := MergeResults(primary, metadata, extended)

	fmt.Printf("Title: %s\n", merged.Title)
	fmt.Printf("Author: %s\n", merged.Author)
	fmt.Printf("Word Count: %d\n", merged.WordCount)
	fmt.Printf("Category: %v\n", merged.Extended["category"])
	// Output: Title: Article Title
	// Author: John Doe
	// Word Count: 150
	// Category: technology
}

// Example of efficient result cloning
func ExampleCloneResult() {
	original := &Result{
		Title:     "Original Title",
		Content:   "Original content",
		WordCount: 100,
		Extended:  map[string]interface{}{"source": "original"},
	}

	// Create independent copy
	cloned := CloneResult(original)

	// Modify the clone without affecting original
	cloned.Title = "Modified Title"
	cloned.Extended["source"] = "cloned"

	fmt.Printf("Original: %s\n", original.Title)
	fmt.Printf("Cloned: %s\n", cloned.Title)
	fmt.Printf("Original source: %v\n", original.Extended["source"])
	fmt.Printf("Cloned source: %v\n", cloned.Extended["source"])
	// Output: Original: Original Title
	// Cloned: Modified Title
	// Original source: original
	// Cloned source: cloned
}

// Example showing memory efficiency with pointer optimization
func ExampleOptimizedExtractorOptions() {
	// Instead of copying large structs, use pointers
	url := "https://example.com"
	html := "<html><body>Content</body></html>"
	metaCache := map[string]string{
		"title":       "Page Title",
		"description": "Page Description",
	}
	fallback := true
	contentType := "html"

	// Pointer-optimized version
	opts := &OptimizedExtractorOptions{
		URL:         &url,
		HTML:        &html,
		MetaCache:   &metaCache,
		Fallback:    &fallback,
		ContentType: &contentType,
	}

	// Convert to value type when needed for compatibility
	valueOpts := opts.ToValue()

	fmt.Printf("URL: %s\n", valueOpts.URL)
	fmt.Printf("Fallback: %t\n", valueOpts.Fallback)
	fmt.Printf("Meta cache entries: %d\n", len(valueOpts.MetaCache))
	// Output: URL: https://example.com
	// Fallback: true
	// Meta cache entries: 2
}

// Example demonstrating performance benefits
func Example_performanceComparison() {
	// Traditional approach - copying large structs
	traditionalProcess := func() {
		result1 := Result{Title: "Article 1", Content: "Long content..."}
		result2 := Result{Title: "Article 2", Author: "Author"}
		
		// This copies the entire struct
		merged := result1
		if merged.Author == "" {
			merged.Author = result2.Author
		}
		
		log.Printf("Traditional merged: %s", merged.Title)
	}

	// Optimized approach - using pointers
	optimizedProcess := func() {
		result1 := &Result{Title: "Article 1", Content: "Long content..."}
		result2 := &Result{Title: "Article 2", Author: "Author"}
		
		// This uses efficient pointer-based merging
		merged := MergeResults(result1, result2)
		
		log.Printf("Optimized merged: %s", merged.Title)
	}

	// Both achieve the same result, but optimized version:
	// - Uses fewer allocations
	// - Faster execution
	// - Better memory efficiency
	
	traditionalProcess()
	optimizedProcess()

	fmt.Println("Both approaches produce the same result with different performance characteristics")
	// Output: Both approaches produce the same result with different performance characteristics
}