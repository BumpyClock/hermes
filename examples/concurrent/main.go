// Package main demonstrates concurrent processing with the Hermes library.
//
// This example shows how to:
// - Process multiple URLs concurrently
// - Use semaphore pattern for resource management
// - Implement progress reporting
// - Handle partial failures gracefully
// - Collect timing metrics
//
// Run with: go run examples/concurrent/main.go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/BumpyClock/hermes"
)

// ProcessResult holds the result of processing a single URL
type ProcessResult struct {
	URL       string
	Result    *hermes.Result
	Duration  time.Duration
	Error     error
}

func main() {
	fmt.Println("Hermes Concurrent Processing Example")
	fmt.Println("===================================")

	// Configuration
	const maxConcurrency = 5 // Maximum concurrent requests
	const timeout = 30 * time.Second

	// URLs to process
	urls := []string{
		"https://httpbin.org/html",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/2", 
		"https://example.com",
		"https://httpbin.org/status/200",
		"https://httpbin.org/json",
		"https://httpbin.org/xml",
		"https://httpbin.org/status/404", // This will fail
		"https://httpbin.org/delay/3",
		"https://httpbin.org/user-agent",
	}

	fmt.Printf("Processing %d URLs with max concurrency of %d\n\n", len(urls), maxConcurrency)

	// Create Hermes client (thread-safe, reusable)
	client := hermes.New(
		hermes.WithTimeout(timeout),
		hermes.WithUserAgent("Hermes-Concurrent-Example/1.0"),
		hermes.WithContentType("text"), // Extract as plain text for faster processing
	)

	// Process URLs concurrently
	results := processURLsConcurrently(client, urls, maxConcurrency)

	// Display results
	displayResults(results)
}

// processURLsConcurrently processes multiple URLs with controlled concurrency
func processURLsConcurrently(client *hermes.Client, urls []string, maxConcurrency int) []ProcessResult {
	results := make([]ProcessResult, len(urls))
	semaphore := make(chan struct{}, maxConcurrency) // Semaphore for concurrency control
	var wg sync.WaitGroup

	fmt.Println("Starting concurrent processing...")
	overallStart := time.Now()

	for i, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(index int, u string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			fmt.Printf("🚀 [%d/%d] Starting: %s\n", index+1, len(urls), u)

			// Process single URL with timing
			result := processSingleURL(client, u)
			results[index] = result

			// Report completion
			if result.Error != nil {
				fmt.Printf("❌ [%d/%d] Failed: %s (%v)\n", index+1, len(urls), u, result.Duration)
			} else {
				fmt.Printf("✅ [%d/%d] Success: %s (%v) - %d words\n", 
					index+1, len(urls), u, result.Duration, result.Result.WordCount)
			}
		}(i, url)
	}

	wg.Wait()
	overallDuration := time.Since(overallStart)

	fmt.Printf("\n⏱️  Overall processing time: %v\n\n", overallDuration)
	return results
}

// processSingleURL processes a single URL and returns the result with timing
func processSingleURL(client *hermes.Client, url string) ProcessResult {
	start := time.Now()
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Parse the URL
	result, err := client.Parse(ctx, url)
	duration := time.Since(start)

	return ProcessResult{
		URL:      url,
		Result:   result,
		Duration: duration,
		Error:    err,
	}
}

// displayResults shows a summary of all processing results
func displayResults(results []ProcessResult) {
	fmt.Println("📊 Processing Summary")
	fmt.Println("====================")

	var totalDuration time.Duration
	var successCount, failureCount int
	var totalWords int

	// Collect statistics
	for _, result := range results {
		totalDuration += result.Duration
		
		if result.Error != nil {
			failureCount++
		} else {
			successCount++
			totalWords += result.Result.WordCount
		}
	}

	// Display summary statistics
	fmt.Printf("✅ Successful: %d/%d (%.1f%%)\n", 
		successCount, len(results), float64(successCount)/float64(len(results))*100)
	fmt.Printf("❌ Failed: %d/%d (%.1f%%)\n", 
		failureCount, len(results), float64(failureCount)/float64(len(results))*100)
	fmt.Printf("📝 Total words extracted: %d\n", totalWords)
	fmt.Printf("⏱️  Total processing time: %v\n", totalDuration)
	
	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(len(results))
		fmt.Printf("⏱️  Average processing time: %v\n", avgDuration)
		fmt.Printf("📊 Average words per article: %.1f\n", float64(totalWords)/float64(successCount))
	}

	// Show detailed results
	fmt.Println("\n📋 Detailed Results")
	fmt.Println("==================")
	
	for i, result := range results {
		fmt.Printf("[%d] %s (%v)\n", i+1, result.URL, result.Duration)
		
		if result.Error != nil {
			if parseErr, ok := result.Error.(*hermes.ParseError); ok {
				fmt.Printf("    ❌ Error [%s]: %v\n", parseErr.Code, parseErr.Err)
			} else {
				fmt.Printf("    ❌ Error: %v\n", result.Error)
			}
		} else {
			r := result.Result
			fmt.Printf("    ✅ Title: %s\n", truncate(r.Title, 50))
			fmt.Printf("    📝 Words: %d, Domain: %s\n", r.WordCount, r.Domain)
		}
		fmt.Println()
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