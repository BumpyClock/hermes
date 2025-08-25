package main

import (
	"fmt"
	"log"
	
	"github.com/BumpyClock/hermes/internal/parser"
)

func main() {
	fmt.Println("🔍 Testing Custom Extractor System")
	fmt.Println("==================================")
	
	p := parser.New()
	
	// Test sites that should use custom extractors
	testSites := []struct {
		name     string
		url      string
		testHTML string
	}{
		{
			name: "NYTimes",
			url:  "https://www.nytimes.com/test",
			testHTML: `<html><head><title>Test</title></head><body>
				<h1 data-testid="headline">NYTimes Custom Test</h1>
				<div class="g-blocks"><p>Content from NYTimes custom extractor</p></div>
			</body></html>`,
		},
		{
			name: "CNN", 
			url:  "https://www.cnn.com/test",
			testHTML: `<html><head><title>Test</title></head><body>
				<h1 class="headline">CNN Custom Test</h1>
				<div class="l-container"><p>Content from CNN custom extractor</p></div>
			</body></html>`,
		},
		{
			name: "Medium",
			url:  "https://medium.com/test",
			testHTML: `<html><head><title>Test</title></head><body>
				<h1>Medium Custom Test</h1>
				<article><p>Content from Medium custom extractor</p></article>
			</body></html>`,
		},
		{
			name: "Unknown Site (Generic)",
			url:  "https://unknown-site.com/test", 
			testHTML: `<html><head><title>Generic Test</title></head><body>
				<h1>Generic Site Test</h1>
				<p>Should use generic extractor</p>
			</body></html>`,
		},
	}
	
	for i, test := range testSites {
		fmt.Printf("\n%d. Testing %s\n", i+1, test.name)
		fmt.Printf("   URL: %s\n", test.url)
		
		result, err := p.ParseHTML(test.testHTML, test.url, &parser.ParserOptions{})
		if err != nil {
			log.Printf("   ❌ Error: %v", err)
			continue
		}
		
		// Check if custom extractor was used
		if result.ExtractorUsed != "" {
			fmt.Printf("   ✅ ExtractorUsed: %s\n", result.ExtractorUsed)
		} else {
			fmt.Printf("   📝 ExtractorUsed: (generic/fallback)\n")
		}
		
		fmt.Printf("   📄 Title: %s\n", result.Title)
		fmt.Printf("   📊 Content Length: %d chars\n", len(result.Content))
	}
	
	fmt.Println("\n🎯 Summary:")
	fmt.Println("- Sites with 'custom:domain' used custom extractors")
	fmt.Println("- Sites with empty ExtractorUsed used generic extractors")
	fmt.Println("- Debug messages show domain matching logic")
}