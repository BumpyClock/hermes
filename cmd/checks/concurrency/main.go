package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/BumpyClock/parser-go/pkg/parser"
)

func main() {
	testHTML := `<!DOCTYPE html>
<html>
<head>
	<title>Test Article</title>
	<meta name="author" content="John Doe">
	<meta property="article:published_time" content="2023-01-15T10:30:00Z">
</head>
<body>
	<article>
		<h1>Test Article Title</h1>
		<p>This is the main content of the article with some meaningful text to extract.</p>
		<p>Second paragraph with more content for better extraction results.</p>
		<p>Third paragraph to make the content substantial enough for proper parsing.</p>
	</article>
</body>
</html>`

	// Test concurrent extraction performance
	fmt.Println("Testing concurrent extraction performance...")
	
	numGoroutines := 100
	numExtractionsPerGoroutine := 10
	
	var wg sync.WaitGroup
	start := time.Now()
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			p := parser.New()
			
			for j := 0; j < numExtractionsPerGoroutine; j++ {
				_, err := p.ParseHTML(testHTML, fmt.Sprintf("https://example.com/test-%d-%d", id, j), &parser.ParserOptions{
					ContentType: "html",
				})
				if err != nil {
					fmt.Printf("Error in goroutine %d, iteration %d: %v\n", id, j, err)
				}
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(start)
	
	totalExtractions := numGoroutines * numExtractionsPerGoroutine
	avgPerExtraction := duration / time.Duration(totalExtractions)
	extractionsPerSecond := float64(totalExtractions) / duration.Seconds()
	
	fmt.Printf("Results:\n")
	fmt.Printf("  Total extractions: %d\n", totalExtractions)
	fmt.Printf("  Total time: %v\n", duration)
	fmt.Printf("  Average per extraction: %v\n", avgPerExtraction)
	fmt.Printf("  Extractions per second: %.2f\n", extractionsPerSecond)
	fmt.Printf("  Concurrent goroutines: %d\n", numGoroutines)
}