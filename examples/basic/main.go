// Package main demonstrates basic usage of the Hermes web content extraction library.
//
// This example shows how to:
// - Create a basic Hermes client
// - Parse a URL with context
// - Handle errors gracefully
// - Access extracted content fields
//
// Run with: go run examples/basic/main.go
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BumpyClock/hermes"
)

func main() {
	fmt.Println("Hermes Basic Example")
	fmt.Println("===================")

	// Create a client with basic configuration
	client := hermes.New(
		hermes.WithTimeout(30*time.Second),
		hermes.WithUserAgent("Hermes-Example/1.0"),
		hermes.WithContentType("html"), // Extract as HTML
	)

	// Example URLs to try
	urls := []string{
		"https://httpbin.org/html",        // Simple test page
		"https://example.com",             // Basic content
		"https://httpbin.org/status/404",  // Error case
	}

	for i, url := range urls {
		fmt.Printf("\n%d. Parsing: %s\n", i+1, url)
		fmt.Println(strings.Repeat("-", len(url)+20))

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		// Parse the URL
		result, err := client.Parse(ctx, url)
		cancel() // Always cancel context when done

		if err != nil {
			// Handle different types of errors
			if parseErr, ok := err.(*hermes.ParseError); ok {
				fmt.Printf("❌ Parse Error [%s]: %v\n", parseErr.Code, parseErr.Err)
				fmt.Printf("   URL: %s\n", parseErr.URL)
				fmt.Printf("   Operation: %s\n", parseErr.Op)
			} else {
				fmt.Printf("❌ Unexpected error: %v\n", err)
			}
			continue
		}

		// Display extracted content
		fmt.Println("✅ Parse successful!")
		displayResult(result)
	}

	fmt.Println("\n🎉 Example completed!")
}

// displayResult formats and displays the extracted content
func displayResult(result *hermes.Result) {
	fmt.Printf("📰 Title: %s\n", truncate(result.Title, 60))
	fmt.Printf("👤 Author: %s\n", result.Author)
	fmt.Printf("🌐 Domain: %s\n", result.Domain)
	fmt.Printf("📝 Word Count: %d\n", result.WordCount)
	
	if result.DatePublished != nil {
		fmt.Printf("📅 Published: %s\n", result.DatePublished.Format("2006-01-02"))
	}
	
	if result.LeadImageURL != "" {
		fmt.Printf("🖼️  Lead Image: %s\n", truncate(result.LeadImageURL, 50))
	}
	
	if result.Description != "" {
		fmt.Printf("📄 Description: %s\n", truncate(result.Description, 100))
	}
	
	if result.Content != "" {
		fmt.Printf("📖 Content: %s...\n", truncate(result.Content, 200))
	}

	if result.Excerpt != "" {
		fmt.Printf("✂️  Excerpt: %s\n", truncate(result.Excerpt, 150))
	}
}

// truncate shortens a string to the specified length with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}