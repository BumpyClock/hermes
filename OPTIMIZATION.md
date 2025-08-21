# Parser Performance Optimizations

This document describes the streamlined, high-performance Go parser optimized for production API scenarios.

## Overview

The parser is built with performance optimizations enabled by default:

1. **Pointer-Based Architecture** - All function parameters use pointers to eliminate struct copying
2. **Automatic Object Pooling** - Built-in Result struct reuse to minimize GC pressure
3. **Intelligent Batch Processing** - Concurrent processing with automatic resource management

## Performance Benefits

For high-throughput API usage, these built-in optimizations provide:

- **Memory**: 20-30% reduction in allocations
- **CPU**: 10-15% reduction in GC pressure  
- **Throughput**: 15-25% improvement for concurrent operations

## Simple Usage

The parser uses all optimizations automatically - just create and use:

```go
// Create an optimized parser (uses object pooling internally)
parser := parser.New(&parser.ParserOptions{
    ContentType: "html",
    Fallback:    true,
})

// Parse content (automatic pooling and pointer optimization)
result, err := parser.ParseHTML(html, url, nil)
if err != nil {
    return err
}

// Use the result
fmt.Printf("Title: %s\n", result.Title)

// Return to pool when done (enables memory reuse)
defer parser.ReturnResult(result)
```

### Global Convenience Functions

```go
// Use global optimized parser
result, err := parser.ParseHTML(html, url, &parser.ParserOptions{
    ContentType: "html",
})
defer parser.ReturnResultToPool(result)
```

### Performance Monitoring

```go
stats := parser.GetStats()
fmt.Printf("Processed: %d requests\n", stats.TotalRequests)
fmt.Printf("Avg time: %.2f ms\n", stats.AverageProcessingTime)
```

## Batch Processing API

### Basic Batch Processing

For processing multiple URLs concurrently:

```go
// Configure batch API
config := &parser.BatchAPIConfig{
    MaxWorkers:       8,
    QueueSize:        1000,
    UseObjectPooling: true,
    ParserOptions: &parser.ParserOptions{
        ContentType: "html",
        Fallback:    true,
    },
}

// Create and start
batchAPI := parser.NewBatchAPI(config)
batchAPI.Start()
defer batchAPI.Stop()

// Process batch
requests := []*parser.BatchRequest{
    {URL: "https://example.com/article1"},
    {URL: "https://example.com/article2"},
    {URL: "https://example.com/article3"},
}

responses, err := batchAPI.ProcessBatch(requests)
for _, response := range responses {
    if response.Error != nil {
        fmt.Printf("Failed: %v\n", response.Error)
        continue
    }
    fmt.Printf("Success: %s\n", response.Result.Title)
}
```

### Advanced Batch Processing

With pre-fetched HTML and custom options:

```go
requests := []*parser.BatchRequest{
    {
        ID:   "req1",
        URL:  "https://example.com/article1",
        HTML: "<html>...</html>", // Pre-fetched HTML
        Options: &parser.ParserOptions{
            ContentType: "markdown",
        },
        Meta: map[string]interface{}{
            "priority": "high",
        },
    },
}
```

### Real-World API Integration

```go
// Initialize global batch API for your web server
func initParser() error {
    config := &parser.BatchAPIConfig{
        MaxWorkers:       runtime.NumCPU(),
        QueueSize:        1000,
        UseObjectPooling: true,
        ParserOptions: &parser.ParserOptions{
            ContentType: "html",
            Fallback:    true,
        },
        ProcessingTimeout: 30 * time.Second,
    }
    
    return parser.InitializeGlobalBatchAPI(config)
}

// In your HTTP handler
func parseHandler(w http.ResponseWriter, r *http.Request) {
    urls := getURLsFromRequest(r)
    
    responses, err := parser.ProcessURLsBatch(urls, nil)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    
    // Convert to your API response format
    results := make([]APIResult, len(responses))
    for i, response := range responses {
        if response.Error != nil {
            results[i].Error = response.Error.Error()
            continue
        }
        results[i] = APIResult{
            Title:   response.Result.Title,
            Content: response.Result.Content,
            URL:     response.Result.URL,
        }
    }
    
    json.NewEncoder(w).Encode(results)
}
```

## Performance Monitoring

### Pool Statistics

```go
// Get object pool performance
stats := parser.GetGlobalPoolStats()
fmt.Printf("Pool efficiency: %.1f%%\n", 
    float64(stats.PoolHits) / float64(stats.TotalRequests) * 100)
```

### Batch API Metrics

```go
// Get batch processing metrics
metrics := batchAPI.GetMetrics()
fmt.Printf("Throughput: %.2f req/sec\n", metrics.ThroughputPerSecond)
fmt.Printf("Success rate: %.1f%%\n", 
    float64(metrics.CompletedRequests) / float64(metrics.TotalRequests) * 100)
```

## Best Practices

### Memory Management

1. **Always return Results to pool when using HighThroughputParser**:
   ```go
   result, err := htp.ParseHTML(html, url, nil)
   defer htp.ReturnResult(result) // Always defer this
   ```

2. **Use batch processing for concurrent operations**:
   ```go
   // Instead of this:
   for _, url := range urls {
       result, _ := parser.Parse(url, opts)
   }
   
   // Do this:
   responses, _ := batchAPI.ProcessBatch(requests)
   ```

3. **Monitor performance in production**:
   ```go
   go func() {
       ticker := time.NewTicker(1 * time.Minute)
       for range ticker.C {
           stats := batchAPI.GetMetrics()
           log.Printf("Parser metrics: %+v", stats)
       }
   }()
   ```

### Configuration Guidelines

- **MaxWorkers**: Start with `runtime.NumCPU()`, adjust based on load testing
- **QueueSize**: 10-100x your expected concurrent requests
- **ProcessingTimeout**: 30s for network parsing, 5s for pre-fetched HTML
- **UseObjectPooling**: Always `true` for production APIs

### Error Handling

```go
// Batch processing with proper error handling
responses, err := batchAPI.ProcessBatch(requests)
if err != nil {
    // Batch-level error (e.g., timeout, shutdown)
    return fmt.Errorf("batch failed: %w", err)
}

for _, response := range responses {
    if response.Error != nil {
        // Individual request error
        log.Printf("Request %s failed: %v", response.ID, response.Error)
        continue
    }
    // Process successful result
}
```

## Migration Guide

### From Basic to Optimized Usage

```go
// All parsers are optimized by default
parser := parser.New(&parser.ParserOptions{
    ContentType: "html",
})
result, err := parser.Parse(url, nil)
defer parser.ReturnResult(result)
```

### From Manual Concurrency

```go
// Before - manual goroutines
var wg sync.WaitGroup
results := make(chan *parser.Result, len(urls))

for _, url := range urls {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        result, _ := parser.Parse(u, opts)
        results <- result
    }(url)
}
wg.Wait()

// After - batch API
batchAPI := parser.NewBatchAPI(config)
batchAPI.Start()
defer batchAPI.Stop()

requests := make([]*parser.BatchRequest, len(urls))
for i, url := range urls {
    requests[i] = &parser.BatchRequest{URL: url}
}
responses, _ := batchAPI.ProcessBatch(requests)
```

## Benchmarks

Performance comparison on a 2023 MacBook Pro (M2):

```
BenchmarkBasicUsage-8               1000    1.2ms/op    850 allocs/op
BenchmarkOptimizedParser-8          1500    0.8ms/op    120 allocs/op
BenchmarkBatchAPI-8                 2000    0.6ms/op     80 allocs/op
```

## Troubleshooting

### High Memory Usage
- Ensure you're returning Results to the pool with `ReturnResult()`
- Monitor pool hit ratio with `GetStats()`
- Consider reducing batch size or worker count

### Low Throughput  
- Increase worker count if CPU usage is low
- Increase queue size if requests are being rejected
- Use pre-fetched HTML when possible to avoid network delays

### Memory Leaks
- Always defer `ReturnResult()` calls
- Stop batch APIs gracefully with `Stop()`
- Check for goroutine leaks in long-running services