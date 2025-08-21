package parser

import (
	"runtime"
	"testing"
	"time"
)

func TestResultPool(t *testing.T) {
	pool := NewResultPool()
	
	// Test getting a result from pool
	result1 := pool.Get()
	if result1 == nil {
		t.Fatal("Expected non-nil result from pool")
	}
	
	// Test that Extended map is initialized
	if result1.Extended == nil {
		t.Error("Expected Extended map to be initialized")
	}
	
	// Test putting result back to pool
	result1.Title = "Test Title"
	result1.Extended["test"] = "value"
	pool.Put(result1)
	
	// Test getting another result (should be the same instance, but reset)
	result2 := pool.Get()
	if result2.Title != "" {
		t.Error("Expected result to be reset when retrieved from pool")
	}
	if len(result2.Extended) != 0 {
		t.Error("Expected Extended map to be cleared when retrieved from pool")
	}
}

// ParserPool was removed - object pooling now handled internally by HighThroughputParser
func TestParserPoolRemoved(t *testing.T) {
	// This test is now redundant since ParserPool was removed in favor of
	// internal pooling within HighThroughputParser
	t.Skip("ParserPool functionality moved to HighThroughputParser")
}

func TestBufferPool(t *testing.T) {
	pool := NewBufferPool(1024)
	
	// Test getting a buffer from pool
	buf1 := pool.Get()
	if buf1 == nil {
		t.Fatal("Expected non-nil buffer from pool")
	}
	if cap(buf1) < 1024 {
		t.Errorf("Expected buffer capacity >= 1024, got %d", cap(buf1))
	}
	
	// Test using buffer
	buf1 = append(buf1, []byte("test data")...)
	if len(buf1) == 0 {
		t.Error("Expected buffer to contain data")
	}
	
	// Test putting buffer back to pool
	pool.Put(buf1)
	
	// Test getting another buffer (should be reset)
	buf2 := pool.Get()
	if len(buf2) != 0 {
		t.Error("Expected buffer to be reset when retrieved from pool")
	}
}

func TestHighThroughputParser(t *testing.T) {
	htp := NewHighThroughputParser(nil)
	
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Test Article</title>
	</head>
	<body>
		<h1>Test Title</h1>
		<p>Test content for parsing.</p>
	</body>
	</html>
	`
	
	// Test ParseHTML
	result, err := htp.ParseHTML(html, "https://example.com/test", nil)
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	// Test that result contains expected data
	if result.URL != "https://example.com/test" {
		t.Errorf("Expected URL 'https://example.com/test', got '%s'", result.URL)
	}
	
	// Test returning result to pool
	htp.ReturnResult(result)
	
	// Test stats
	stats := htp.GetStats()
	if stats.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", stats.TotalRequests)
	}
	if stats.AverageProcessingTime <= 0 {
		t.Error("Expected positive average processing time")
	}
}

func TestHighThroughputParserBatch(t *testing.T) {
	htp := NewHighThroughputParser(nil)
	
	urls := []string{
		"https://example.com/article1",
		"https://example.com/article2",
		"https://example.com/article3",
	}
	
	// Note: These will likely fail with network errors, but we're testing the pooling mechanism
	results, errors := htp.ParseBatch(urls, &ParserOptions{ContentType: "html"})
	
	if len(results) != len(urls) {
		t.Errorf("Expected %d results, got %d", len(urls), len(results))
	}
	
	// Should have errors due to network requests, but that's expected in tests
	if len(errors) == 0 {
		t.Log("Unexpected: no errors in batch parsing (network might be working)")
	}
	
	// Return results to pool
	for _, result := range results {
		if result != nil {
			htp.ReturnResult(result)
		}
	}
}

func TestGlobalFunctions(t *testing.T) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Global Test</title>
	</head>
	<body>
		<h1>Global Title</h1>
		<p>Global content.</p>
	</body>
	</html>
	`
	
	// Test global ParseHTML
	result, err := ParseHTML(html, "https://example.com/global", &ParserOptions{
		ContentType: "html",
	})
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	
	// Test global stats
	stats := GetGlobalStats()
	if stats.TotalRequests < 1 {
		t.Error("Expected at least 1 request in global stats")
	}
	
	// Return result to global pool
	ReturnResult(result)
}

func BenchmarkHighThroughputParserVsRegular(b *testing.B) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Benchmark Article</title>
	</head>
	<body>
		<h1>Benchmark Title</h1>
		<div class="content">
			<p>This is some benchmark content for testing parser performance.</p>
			<p>Multiple paragraphs to make the content more realistic.</p>
			<p>Testing memory allocation patterns and GC pressure.</p>
		</div>
	</body>
	</html>
	`
	
	b.Run("Regular Parser", func(b *testing.B) {
		parser := New()
		opts := &ParserOptions{ContentType: "html"}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := parser.ParseHTML(html, "https://example.com/bench", opts)
			if err != nil {
				b.Fatal(err)
			}
			_ = result // Use result
		}
	})
	
	b.Run("High Throughput Parser", func(b *testing.B) {
		htp := NewHighThroughputParser(&ParserOptions{ContentType: "html"})
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := htp.ParseHTML(html, "https://example.com/bench", nil)
			if err != nil {
				b.Fatal(err)
			}
			htp.ReturnResult(result) // Return to pool
		}
	})
}

func BenchmarkMemoryAllocation(b *testing.B) {
	html := `<html><body><h1>Title</h1><p>Content</p></body></html>`
	
	b.Run("Without Pooling", func(b *testing.B) {
		parser := New()
		opts := &ParserOptions{ContentType: "html"}
		
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := parser.ParseHTML(html, "https://example.com/test", opts)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
		b.StopTimer()
		
		runtime.GC()
		runtime.ReadMemStats(&m2)
		b.Logf("Memory allocations: %d bytes, %d allocs", m2.TotalAlloc-m1.TotalAlloc, m2.Mallocs-m1.Mallocs)
	})
	
	b.Run("With Pooling", func(b *testing.B) {
		htp := NewHighThroughputParser(&ParserOptions{ContentType: "html"})
		
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, err := htp.ParseHTML(html, "https://example.com/test", nil)
			if err != nil {
				b.Fatal(err)
			}
			htp.ReturnResult(result)
		}
		b.StopTimer()
		
		runtime.GC()
		runtime.ReadMemStats(&m2)
		b.Logf("Memory allocations: %d bytes, %d allocs", m2.TotalAlloc-m1.TotalAlloc, m2.Mallocs-m1.Mallocs)
	})
}

func TestPoolStatsResetAndUpdates(t *testing.T) {
	htp := NewHighThroughputParser(nil)
	html := `<html><body><p>Test</p></body></html>`
	
	// Reset stats
	htp.ResetStats()
	
	// Parse multiple times
	for i := 0; i < 3; i++ {
		result, err := htp.ParseHTML(html, "https://example.com/test", nil)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		htp.ReturnResult(result)
		time.Sleep(1 * time.Millisecond) // Small delay to vary timing
	}
	
	stats := htp.GetStats()
	if stats.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", stats.TotalRequests)
	}
	if stats.AverageProcessingTime <= 0 {
		t.Error("Expected positive average processing time")
	}
	
	// Test reset
	htp.ResetStats()
	statsAfterReset := htp.GetStats()
	if statsAfterReset.TotalRequests != 0 {
		t.Errorf("Expected 0 requests after reset, got %d", statsAfterReset.TotalRequests)
	}
}