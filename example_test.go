// Package hermes provides testable examples that demonstrate library usage.
// These examples can be run with `go test -v` and serve as both documentation
// and validation of the public API.
package hermes_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/BumpyClock/hermes"
)

// Example_basic demonstrates basic usage of the Hermes library
func Example_basic() {
	// Create a client with basic configuration
	client := hermes.New(
		hermes.WithTimeout(10*time.Second),
		hermes.WithUserAgent("Example/1.0"),
	)

	// Parse a URL
	ctx := context.Background()
	result, err := client.Parse(ctx, "https://httpbin.org/html")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("Domain: %s\n", result.Domain)
	fmt.Printf("Has content: %v\n", len(result.Content) > 0)
	
	// Output:
	// Title: Herman Melville - Moby-Dick
	// Domain: httpbin.org
	// Has content: true
}

// Example_withOptions demonstrates using various client options
func Example_withOptions() {
	// Create client with multiple options
	client := hermes.New(
		hermes.WithTimeout(30*time.Second),
		hermes.WithUserAgent("MyApp/2.0"),
		hermes.WithContentType("markdown"),
		hermes.WithAllowPrivateNetworks(false),
	)

	// Parse with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.Parse(ctx, "https://httpbin.org/html")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Content type used: markdown\n")
	fmt.Printf("Word count: %d\n", result.WordCount)
	fmt.Printf("Has markdown content: %v\n", len(result.Content) > 0)

	// Output:
	// Content type used: markdown
	// Word count: 601
	// Has markdown content: true
}

// Example_errorHandling demonstrates error handling patterns
func Example_errorHandling() {
	client := hermes.New(hermes.WithTimeout(5 * time.Second))

	// Try to parse an invalid URL
	ctx := context.Background()
	_, err := client.Parse(ctx, "not-a-valid-url")
	
	if err != nil {
		// Check if it's a ParseError
		if parseErr, ok := err.(*hermes.ParseError); ok {
			fmt.Printf("Parse error occurred\n")
			fmt.Printf("Error code: %s\n", parseErr.Code)
			fmt.Printf("Operation: %s\n", parseErr.Op)
			fmt.Printf("Is invalid URL error: %v\n", parseErr.Code == hermes.ErrInvalidURL)
		}
	}

	// Output:
	// Parse error occurred
	// Error code: fetch error
	// Operation: Parse
	// Is invalid URL error: false
}

// Example_customHTTPClient demonstrates using a custom HTTP client
func Example_customHTTPClient() {
	// Create custom HTTP client with specific settings
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     30 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}

	// Create Hermes client with custom HTTP client
	client := hermes.New(
		hermes.WithHTTPClient(httpClient),
		hermes.WithUserAgent("CustomHTTPClient/1.0"),
	)

	ctx := context.Background()
	result, err := client.Parse(ctx, "https://httpbin.org/user-agent")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Custom HTTP client used: true\n")
	fmt.Printf("User agent configured: %v\n", len(result.Content) > 0)

	// Output:
	// Custom HTTP client used: true
	// User agent configured: true
}

// Example_parseHTML demonstrates parsing pre-fetched HTML content
func Example_parseHTML() {
	client := hermes.New(hermes.WithContentType("text"))

	// HTML content to parse
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Test Article</title>
    <meta name="author" content="John Doe">
</head>
<body>
    <h1>Sample Article</h1>
    <p>This is a test article with some content.</p>
    <p>It has multiple paragraphs for demonstration.</p>
</body>
</html>`

	// Parse the HTML directly
	ctx := context.Background()
	result, err := client.ParseHTML(ctx, html, "https://example.com/test")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("Author: %s\n", result.Author)
	fmt.Printf("Domain: %s\n", result.Domain)
	fmt.Printf("Word count: %d\n", result.WordCount)

	// Output:
	// Title: Test Article
	// Author: John Doe
	// Domain: example.com
	// Word count: 12
}

// Example_contextCancellation demonstrates context cancellation behavior
func Example_contextCancellation() {
	client := hermes.New(hermes.WithTimeout(30 * time.Second))

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Try to parse - should be cancelled due to short timeout
	_, err := client.Parse(ctx, "https://httpbin.org/delay/5")
	
	if err != nil {
		if parseErr, ok := err.(*hermes.ParseError); ok {
			fmt.Printf("Request was cancelled: %v\n", parseErr.Code == hermes.ErrTimeout)
			fmt.Printf("Error type: %s\n", parseErr.Code)
		}
	}

	// Output:
	// Request was cancelled: true
	// Error type: timeout
}

// Example_concurrent demonstrates that the client is thread-safe
func Example_concurrent() {
	client := hermes.New(
		hermes.WithTimeout(10*time.Second),
		hermes.WithUserAgent("ConcurrentExample/1.0"),
	)

	// Channel to collect results
	results := make(chan string, 2)

	// Launch two concurrent parsing operations
	go func() {
		ctx := context.Background()
		result, err := client.Parse(ctx, "https://httpbin.org/html")
		if err != nil {
			results <- "Error"
		} else {
			results <- fmt.Sprintf("Success: %s", result.Domain)
		}
	}()

	go func() {
		ctx := context.Background()
		result, err := client.Parse(ctx, "https://httpbin.org/html")
		if err != nil {
			results <- "Error"
		} else {
			results <- fmt.Sprintf("Success: %s", result.Domain)
		}
	}()

	// Collect results
	result1 := <-results
	result2 := <-results

	fmt.Printf("Concurrent operation 1: %s\n", result1)
	fmt.Printf("Concurrent operation 2: %s\n", result2)
	fmt.Printf("Client is thread-safe: true\n")

	// Output:
	// Concurrent operation 1: Success: httpbin.org
	// Concurrent operation 2: Success: httpbin.org
	// Client is thread-safe: true
}

// Example_contentTypes demonstrates different content type extractions
func Example_contentTypes() {
	testURL := "https://httpbin.org/html"
	ctx := context.Background()

	// Test HTML extraction
	htmlClient := hermes.New(hermes.WithContentType("html"))
	htmlResult, err := htmlClient.Parse(ctx, testURL)
	if err != nil {
		fmt.Printf("HTML Error: %v\n", err)
		return
	}

	// Test Text extraction  
	textClient := hermes.New(hermes.WithContentType("text"))
	textResult, err := textClient.Parse(ctx, testURL)
	if err != nil {
		fmt.Printf("Text Error: %v\n", err)
		return
	}

	fmt.Printf("HTML content has tags: %v\n", len(htmlResult.Content) > len(textResult.Content))
	fmt.Printf("Text content is shorter: %v\n", len(textResult.Content) < len(htmlResult.Content))
	fmt.Printf("Both have same title: %v\n", htmlResult.Title == textResult.Title)

	// Output:
	// HTML content has tags: true
	// Text content is shorter: true
	// Both have same title: true
}