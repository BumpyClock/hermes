package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BumpyClock/parser-go/pkg/parser"
)

func main() {
	// Test real-world fixture files  
	testFiles := []struct {
		filename string
		domain   string
		url      string
	}{
		{"www.nytimes.com.html", "www.nytimes.com", "https://www.nytimes.com/test-article"},
		{"www.cnn.com.html", "www.cnn.com", "https://www.cnn.com/test-article"},
		{"www.theverge.com.html", "www.theverge.com", "https://www.theverge.com/test-article"},
		{"www.wired.com.html", "www.wired.com", "https://www.wired.com/test-article"},
		{"medium.com.html", "medium.com", "https://medium.com/test-article"},
		{"arstechnica.com.html", "arstechnica.com", "https://arstechnica.com/test-article"},
	}

	p := parser.New()
	
	fmt.Println("Testing real-world extraction on fixture files...")
	fmt.Println(strings.Repeat("=", 60))
	
	totalTime := time.Duration(0)
	successCount := 0
	
	for _, test := range testFiles {
		fmt.Printf("\nTesting %s:\n", test.filename)
		
		// Read fixture file
		htmlContent, err := os.ReadFile(fmt.Sprintf("internal/fixtures/%s", test.filename))
		if err != nil {
			fmt.Printf("  ❌ Could not read fixture file: %v\n", err)
			continue
		}
		
		// Parse with timing
		start := time.Now()
		result, err := p.ParseHTML(string(htmlContent), test.url, parser.ParserOptions{
			ContentType: "html",
		})
		duration := time.Since(start)
		totalTime += duration
		
		if err != nil {
			fmt.Printf("  ❌ Parse error: %v\n", err)
			continue
		}
		
		if result == nil {
			fmt.Printf("  ❌ No result returned\n")
			continue
		}
		
		successCount++
		
		// Display results
		fmt.Printf("  ✅ Success (took %v)\n", duration)
		fmt.Printf("     Title: %s\n", truncate(result.Title, 80))
		fmt.Printf("     Author: %s\n", truncate(result.Author, 40))
		fmt.Printf("     Domain: %s\n", result.Domain)
		fmt.Printf("     Word Count: %d\n", result.WordCount)
		fmt.Printf("     Content Length: %d chars\n", len(result.Content))
		
		if result.DatePublished != nil {
			fmt.Printf("     Date: %v\n", result.DatePublished.Format("2006-01-02"))
		}
		
		// Show first 200 chars of content
		if len(result.Content) > 0 {
			fmt.Printf("     Content Preview: %s...\n", truncate(stripHTML(result.Content), 150))
		}
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Summary:\n")
	fmt.Printf("  Files tested: %d\n", len(testFiles))
	fmt.Printf("  Successful extractions: %d\n", successCount)
	fmt.Printf("  Success rate: %.1f%%\n", float64(successCount)/float64(len(testFiles))*100)
	fmt.Printf("  Total time: %v\n", totalTime)
	if successCount > 0 {
		fmt.Printf("  Average per extraction: %v\n", totalTime/time.Duration(successCount))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Simple HTML tag stripper for preview
func stripHTML(s string) string {
	inTag := false
	result := ""
	for _, r := range s {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result += string(r)
		}
	}
	return result
}