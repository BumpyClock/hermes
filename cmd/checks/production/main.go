package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
	fmt.Println("Production Readiness Assessment")
	fmt.Println(strings.Repeat("=", 50))
	
	p := parser.New()
	
	// Test 1: Error Handling - Invalid URL
	fmt.Println("\n1. Testing invalid URL handling:")
	result, err := p.ParseHTML("", "not-a-url", &parser.ParserOptions{})
	if err != nil {
		fmt.Printf("   ✅ Properly handles invalid URLs: %v\n", err)
	} else if result != nil && result.Error {
		fmt.Printf("   ✅ Returns error result for invalid URLs: %s\n", result.Message)
	}
	
	// Test 2: Error Handling - Malformed HTML
	fmt.Println("\n2. Testing malformed HTML handling:")
	malformedHTML := `<html><head><title>Test</><body><p>Unclosed tags<div>More content`
	result, err = p.ParseHTML(malformedHTML, "https://example.com/test", &parser.ParserOptions{})
	if err != nil {
		fmt.Printf("   ❌ Failed on malformed HTML: %v\n", err)
	} else if result != nil {
		fmt.Printf("   ✅ Gracefully handles malformed HTML - extracted title: '%s'\n", result.Title)
	}
	
	// Test 3: Empty content handling
	fmt.Println("\n3. Testing empty content handling:")
	emptyHTML := `<html><head><title>Empty</title></head><body></body></html>`
	result, err = p.ParseHTML(emptyHTML, "https://example.com/empty", &parser.ParserOptions{})
	if err != nil {
		fmt.Printf("   ❌ Failed on empty content: %v\n", err)
	} else if result != nil {
		fmt.Printf("   ✅ Handles empty content - title: '%s', content length: %d\n", result.Title, len(result.Content))
	}
	
	// Test 4: Large document handling
	fmt.Println("\n4. Testing large document handling:")
	largeContent := strings.Repeat("<p>Large paragraph content with lots of text to test memory usage and performance on bigger documents. ", 1000)
	largeHTML := fmt.Sprintf(`<html><head><title>Large Document</title></head><body><article>%s</article></body></html>`, largeContent)
	
	start := time.Now()
	result, err = p.ParseHTML(largeHTML, "https://example.com/large", &parser.ParserOptions{})
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("   ❌ Failed on large document: %v\n", err)
	} else if result != nil {
		fmt.Printf("   ✅ Handles large documents (%.1fKB) in %v - word count: %d\n", 
			float64(len(largeHTML))/1024, duration, result.WordCount)
	}
	
	// Test 5: Content type handling
	fmt.Println("\n5. Testing content type variations:")
	testHTML := `<html><head><title>Content Types</title></head><body><article><p>Test content.</p></article></body></html>`
	
	contentTypes := []string{"html", "markdown", "text"}
	for _, ct := range contentTypes {
		result, err = p.ParseHTML(testHTML, "https://example.com/content", &parser.ParserOptions{
			ContentType: ct,
		})
		if err != nil {
			fmt.Printf("   ❌ Failed with content type %s: %v\n", ct, err)
		} else if result != nil {
			fmt.Printf("   ✅ Content type '%s': content length %d chars\n", ct, len(result.Content))
		}
	}
	
	// Test 6: Configuration options
	fmt.Println("\n6. Testing configuration options:")
	result, err = p.ParseHTML(testHTML, "https://example.com/config", &parser.ParserOptions{
		FetchAllPages: false,
		Fallback:      true,
		ContentType:   "html",
		Headers:       map[string]string{"User-Agent": "Test-Parser"},
	})
	if err != nil {
		fmt.Printf("   ❌ Failed with custom options: %v\n", err)
	} else if result != nil {
		fmt.Printf("   ✅ Custom options work - extracted title: '%s'\n", result.Title)
	}
	
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("Production readiness assessment complete!")
}