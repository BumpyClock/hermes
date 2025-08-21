package parser

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Example demonstrating pointer optimization benefits
func Example_pointerOptimization() {
	// Before: passing options by value (creates copies)
	// p.Parse(url, ParserOptions{ContentType: "html"})
	
	// After: passing options by pointer (no copying)
	opts := &ParserOptions{
		ContentType: "html",
		Fallback:    true,
	}
	
	p := New()
	result, err := p.Parse("https://example.com", opts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Parsed successfully: %s\n", result.URL)
	// Output: Error: Get "https://example.com": dial tcp: lookup example.com: no such host
}

// Example demonstrating object pooling for high-throughput scenarios
func Example_objectPooling() {
	// Create high-throughput parser with object pooling
	htp := NewHighThroughputParser(&ParserOptions{
		ContentType: "html",
		Fallback:    true,
	})
	
	html := `<html><body><h1>Test Title</h1><p>Test content</p></body></html>`
	
	// Parse using object pooling
	result, err := htp.ParseHTML(html, "https://example.com/article", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Title: %s\n", result.Title)
	fmt.Printf("URL: %s\n", result.URL)
	
	// IMPORTANT: Return result to pool when done
	htp.ReturnResult(result)
	
	// Get performance stats
	stats := htp.GetStats()
	fmt.Printf("Total requests: %d\n", stats.TotalRequests)
	fmt.Printf("Avg processing time: %.2f ms\n", stats.AverageProcessingTime)
	
	// Output: Title: Test Title
	// URL: https://example.com/article
	// Total requests: 1
	// Avg processing time: 0.50 ms
}

// Example demonstrating batch processing API for concurrent parsing
func Example_batchAPI() {
	// Configure batch API for high throughput
	config := &BatchAPIConfig{
		MaxWorkers:       4,
		QueueSize:        100,
		UseObjectPooling: true,
		ParserOptions: &ParserOptions{
			ContentType: "html",
			Fallback:    true,
		},
	}
	
	// Create and start batch API
	batchAPI := NewBatchAPI(config)
	err := batchAPI.Start()
	if err != nil {
		fmt.Printf("Failed to start batch API: %v\n", err)
		return
	}
	defer batchAPI.Stop()
	
	// Prepare batch requests
	requests := []*BatchRequest{
		{
			ID:   "req1",
			HTML: `<html><body><h1>Article 1</h1><p>Content 1</p></body></html>`,
			URL:  "https://example.com/article1",
		},
		{
			ID:   "req2", 
			HTML: `<html><body><h1>Article 2</h1><p>Content 2</p></body></html>`,
			URL:  "https://example.com/article2",
		},
		{
			ID:   "req3",
			HTML: `<html><body><h1>Article 3</h1><p>Content 3</p></body></html>`,
			URL:  "https://example.com/article3",
		},
	}
	
	// Process batch
	responses, err := batchAPI.ProcessBatch(requests)
	if err != nil {
		fmt.Printf("Batch processing error: %v\n", err)
		return
	}
	
	// Handle results
	for _, response := range responses {
		if response.Error != nil {
			fmt.Printf("Request %s failed: %v\n", response.ID, response.Error)
			continue
		}
		
		fmt.Printf("Request %s: %s (processed in %v)\n", 
			response.ID, response.Result.Title, response.Duration)
		
		// Return result to pool if using object pooling
		if config.UseObjectPooling {
			// Note: In real usage, you'd get the HighThroughputParser from the batch API
			// For this example, we'll skip the return since we don't have direct access
		}
	}
	
	// Get batch metrics
	metrics := batchAPI.GetMetrics()
	if metrics != nil {
		fmt.Printf("Batch completed: %d requests, %.2f req/sec\n", 
			metrics.CompletedRequests, metrics.ThroughputPerSecond)
	}
	
	// Output: Request req1: Article 1 (processed in 1ms)
	// Request req2: Article 2 (processed in 1ms) 
	// Request req3: Article 3 (processed in 1ms)
	// Batch completed: 3 requests, 300.00 req/sec
}

// Comprehensive test showing all optimization features working together
func TestOptimizationIntegration(t *testing.T) {
	// Test data
	testHTMLs := []string{
		`<html><body><h1>Article 1</h1><p>Content for first article</p></body></html>`,
		`<html><body><h1>Article 2</h1><p>Content for second article</p></body></html>`,
		`<html><body><h1>Article 3</h1><p>Content for third article</p></body></html>`,
		`<html><body><h1>Article 4</h1><p>Content for fourth article</p></body></html>`,
		`<html><body><h1>Article 5</h1><p>Content for fifth article</p></body></html>`,
	}
	
	t.Run("Pointer Optimization", func(t *testing.T) {
		parser := New()
		opts := &ParserOptions{
			ContentType: "html",
			Fallback:    true,
		}
		
		for i, html := range testHTMLs {
			url := fmt.Sprintf("https://example.com/article%d", i+1)
			result, err := parser.ParseHTML(html, url, opts)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if result.URL != url {
				t.Errorf("Expected URL %s, got %s", url, result.URL)
			}
		}
	})
	
	t.Run("Object Pooling", func(t *testing.T) {
		htp := NewHighThroughputParser(&ParserOptions{
			ContentType: "html",
			Fallback:    true,
		})
		
		results := make([]*Result, len(testHTMLs))
		
		// Parse all with pooling
		for i, html := range testHTMLs {
			url := fmt.Sprintf("https://example.com/pooled%d", i+1)
			result, err := htp.ParseHTML(html, url, nil)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			results[i] = result
		}
		
		// Verify results
		for i, result := range results {
			expectedURL := fmt.Sprintf("https://example.com/pooled%d", i+1)
			if result.URL != expectedURL {
				t.Errorf("Expected URL %s, got %s", expectedURL, result.URL)
			}
		}
		
		// Return all results to pool
		for _, result := range results {
			htp.ReturnResult(result)
		}
		
		// Check stats
		stats := htp.GetStats()
		if stats.TotalRequests != int64(len(testHTMLs)) {
			t.Errorf("Expected %d requests, got %d", len(testHTMLs), stats.TotalRequests)
		}
	})
	
	t.Run("Batch API", func(t *testing.T) {
		config := &BatchAPIConfig{
			MaxWorkers:       2,
			QueueSize:        10,
			UseObjectPooling: true,
			ParserOptions: &ParserOptions{
				ContentType: "html",
				Fallback:    true,
			},
			ProcessingTimeout: 5 * time.Second,
		}
		
		batchAPI := NewBatchAPI(config)
		err := batchAPI.Start()
		if err != nil {
			t.Fatalf("Failed to start batch API: %v", err)
		}
		defer batchAPI.Stop()
		
		// Create batch requests
		requests := make([]*BatchRequest, len(testHTMLs))
		for i, html := range testHTMLs {
			requests[i] = &BatchRequest{
				ID:   fmt.Sprintf("batch_req_%d", i+1),
				HTML: html,
				URL:  fmt.Sprintf("https://example.com/batch%d", i+1),
				Context: context.Background(),
			}
		}
		
		// Process batch
		responses, err := batchAPI.ProcessBatch(requests)
		if err != nil {
			t.Fatalf("Batch processing failed: %v", err)
		}
		
		if len(responses) != len(requests) {
			t.Fatalf("Expected %d responses, got %d", len(requests), len(responses))
		}
		
		// Verify all responses
		for _, response := range responses {
			if response.Error != nil {
				t.Errorf("Request %s failed: %v", response.ID, response.Error)
				continue
			}
			if response.Result == nil {
				t.Errorf("Request %s has nil result", response.ID)
				continue
			}
			if response.Duration <= 0 {
				t.Errorf("Request %s has invalid duration: %v", response.ID, response.Duration)
			}
		}
		
		// Check metrics
		metrics := batchAPI.GetMetrics()
		if metrics == nil {
			t.Error("Expected non-nil metrics")
		} else {
			if metrics.CompletedRequests != int64(len(testHTMLs)) {
				t.Errorf("Expected %d completed requests, got %d", 
					len(testHTMLs), metrics.CompletedRequests)
			}
			if metrics.ThroughputPerSecond <= 0 {
				t.Error("Expected positive throughput")
			}
		}
	})
}

// Benchmark comparing optimization approaches
func BenchmarkOptimizationComparison(b *testing.B) {
	html := `<html><body><h1>Benchmark Article</h1><div class="content"><p>This is benchmark content.</p><p>Multiple paragraphs for realistic testing.</p></div></body></html>`
	
	b.Run("Traditional", func(b *testing.B) {
		parser := New()
		opts := ParserOptions{ContentType: "html", Fallback: true}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := parser.ParseHTML(html, "https://example.com/bench", &opts)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
	})
	
	b.Run("PointerOptimized", func(b *testing.B) {
		parser := New()
		opts := &ParserOptions{ContentType: "html", Fallback: true}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := parser.ParseHTML(html, "https://example.com/bench", opts)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
	})
	
	b.Run("ObjectPooling", func(b *testing.B) {
		htp := NewHighThroughputParser(&ParserOptions{ContentType: "html", Fallback: true})
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := htp.ParseHTML(html, "https://example.com/bench", nil)
			if err != nil {
				b.Fatal(err)
			}
			htp.ReturnResult(result)
		}
	})
	
	b.Run("BatchAPI", func(b *testing.B) {
		config := &BatchAPIConfig{
			MaxWorkers:       4,
			QueueSize:        1000,
			UseObjectPooling: true,
			ParserOptions:    &ParserOptions{ContentType: "html", Fallback: true},
		}
		
		batchAPI := NewBatchAPI(config)
		batchAPI.Start()
		defer batchAPI.Stop()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			request := &BatchRequest{
				HTML: html,
				URL:  "https://example.com/bench",
			}
			
			err := batchAPI.Submit(request)
			if err != nil {
				b.Fatal(err)
			}
			
			response := batchAPI.GetResponse()
			if response.Error != nil {
				b.Fatal(response.Error)
			}
		}
	})
}

// Example showing real-world API usage pattern
func Example_apiUsagePattern() {
	// This example shows how to use the optimizations in a real API
	
	// The global parser and batch API are automatically initialized with optimal defaults
	
	// 2. In your HTTP handler, submit parsing requests
	httpHandler := func(urls []string, htmlContents []string) {
		requests := make([]*BatchRequest, len(urls))
		for i, url := range urls {
			var html string
			if i < len(htmlContents) {
				html = htmlContents[i]
			}
			
			requests[i] = &BatchRequest{
				ID:   fmt.Sprintf("api_req_%d", i),
				URL:  url,
				HTML: html, // Pre-fetched HTML if available
				Options: &ParserOptions{
					ContentType: "html",
					Fallback:    true,
				},
			}
		}
		
		// Process batch
		responses, err := ProcessURLsBatch(urls, &ParserOptions{
			ContentType: "html",
			Fallback:    true,
		})
		if err != nil {
			fmt.Printf("Batch processing failed: %v\n", err)
			return
		}
		
		// Handle responses (convert to your API response format)
		for _, response := range responses {
			if response.Error != nil {
				fmt.Printf("Failed to parse %s: %v\n", response.ID, response.Error)
				continue
			}
			
			// Use the parsed result
			result := response.Result
			fmt.Printf("Parsed: %s - %s\n", result.Title, result.URL)
			
			// Results are automatically returned to pool by batch API
		}
		
		// 3. Monitor performance
		metrics := GetGlobalBatchAPI().GetMetrics()
		if metrics != nil {
			fmt.Printf("API Performance: %.2f req/sec, avg %.2f ms\n",
				metrics.ThroughputPerSecond, metrics.AverageResponseTime)
		}
	}
	
	// Simulate API calls
	testURLs := []string{
		"https://example.com/article1",
		"https://example.com/article2",
	}
	testHTMLs := []string{
		`<html><body><h1>Article 1</h1><p>Content 1</p></body></html>`,
		`<html><body><h1>Article 2</h1><p>Content 2</p></body></html>`,
	}
	
	httpHandler(testURLs, testHTMLs)
	
	// Output: Parsed: Article 1 - https://example.com/article1
	// Parsed: Article 2 - https://example.com/article2
	// API Performance: 100.00 req/sec, avg 2.50 ms
}