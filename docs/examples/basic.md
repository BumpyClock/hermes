# Basic Examples

This document provides practical examples of using Hermes for common content extraction scenarios.

## Table of Contents

- [Simple Content Extraction](#simple-content-extraction)
- [Configuration Examples](#configuration-examples)
- [Output Format Examples](#output-format-examples)
- [Error Handling Examples](#error-handling-examples)
- [Batch Processing Examples](#batch-processing-examples)
- [Real-World Use Cases](#real-world-use-cases)

## Simple Content Extraction

### Extract Article from URL

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    // Create a parser with default settings
    p := parser.New()
    
    // Extract content from a news article
    result, err := p.Parse("https://www.theguardian.com/technology/ai", nil)
    if err != nil {
        log.Fatal("Failed to parse URL:", err)
    }
    
    if result.IsError() {
        log.Fatal("Extraction failed:", result.Message)
    }
    
    // Display extracted information
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Word Count: %d\n", result.WordCount)
    fmt.Printf("Domain: %s\n", result.Domain)
    
    if result.DatePublished != nil {
        fmt.Printf("Published: %s\n", result.DatePublished.Format("January 2, 2006"))
    }
    
    if result.LeadImageURL != "" {
        fmt.Printf("Lead Image: %s\n", result.LeadImageURL)
    }
    
    // Print first 200 characters of content
    if len(result.Content) > 200 {
        fmt.Printf("Content Preview: %s...\n", result.Content[:200])
    } else {
        fmt.Printf("Content: %s\n", result.Content)
    }
}
```

### Extract from HTML String

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    htmlContent := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Sample Article</title>
        <meta name="author" content="Jane Doe">
        <meta property="article:published_time" content="2024-01-15T10:30:00Z">
    </head>
    <body>
        <article>
            <h1>Breaking: New Technology Breakthrough</h1>
            <p class="byline">By Jane Doe</p>
            <time datetime="2024-01-15T10:30:00Z">January 15, 2024</time>
            
            <div class="article-content">
                <p>Scientists have made a significant breakthrough in quantum computing technology...</p>
                <p>The research, published in Nature, demonstrates a new approach to quantum error correction.</p>
                <h2>Key Findings</h2>
                <ul>
                    <li>50% improvement in error correction rates</li>
                    <li>Scalable to 1000+ qubit systems</li>
                    <li>Operates at higher temperatures</li>
                </ul>
                <p>This development could accelerate the timeline for practical quantum computers.</p>
            </div>
        </article>
    </body>
    </html>`
    
    p := parser.New()
    
    // Parse HTML directly
    result, err := p.ParseHTML(htmlContent, "https://example.com/article", nil)
    if err != nil {
        log.Fatal("Failed to parse HTML:", err)
    }
    
    if result.IsError() {
        log.Fatal("Extraction failed:", result.Message)
    }
    
    fmt.Printf("Extracted Title: %s\n", result.Title)
    fmt.Printf("Extracted Author: %s\n", result.Author)
    fmt.Printf("Word Count: %d\n", result.WordCount)
    fmt.Printf("Content:\n%s\n", result.Content)
}
```

## Configuration Examples

### Custom Headers and User Agent

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    // Configure parser with custom headers
    opts := &parser.ParserOptions{
        ContentType: "html",
        Headers: map[string]string{
            "User-Agent": "MyNewsBot/1.0 (+https://example.com/bot)",
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
            "Accept-Language": "en-US,en;q=0.5",
            "Accept-Encoding": "gzip, deflate",
            "DNT": "1",
            "Connection": "keep-alive",
            "Pragma": "no-cache",
            "Cache-Control": "no-cache",
        },
    }
    
    p := parser.New(opts)
    
    // Parse with custom headers
    result, err := p.Parse("https://www.reuters.com/technology/", opts)
    if err != nil {
        log.Fatal("Failed to parse:", err)
    }
    
    if result.IsError() {
        log.Fatal("Extraction failed:", result.Message)
    }
    
    fmt.Printf("Successfully extracted: %s\n", result.Title)
}
```

### Disable Multi-page Fetching

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    // Configure for single-page extraction (faster)
    opts := &parser.ParserOptions{
        FetchAllPages: false, // Only extract the first page
        Fallback:      true,
        ContentType:   "markdown",
    }
    
    p := parser.New(opts)
    
    start := time.Now()
    result, err := p.Parse("https://www.wired.com/story/ai-chatbots/", opts)
    duration := time.Since(start)
    
    if err != nil {
        log.Fatal("Failed to parse:", err)
    }
    
    if result.IsError() {
        log.Fatal("Extraction failed:", result.Message)
    }
    
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Extraction time: %v\n", duration)
    fmt.Printf("Pages extracted: %d of %d\n", result.RenderedPages, result.TotalPages)
    
    if result.NextPageURL != "" {
        fmt.Printf("Next page available: %s\n", result.NextPageURL)
    }
}
```

### Extended Field Extraction

```go
package main

import (
    "fmt"
    "log"
    "strings"
    
    "github.com/BumpyClock/hermes/pkg/parser"
    "github.com/PuerkitoBio/goquery"
)

func main() {
    // Configure parser with extended field extraction
    opts := &parser.ParserOptions{
        ContentType: "html",
        Extend: map[string]parser.ExtractorFunc{
            "reading_time": func(doc *goquery.Document, url string) (interface{}, error) {
                text := doc.Find("article, .article-content, .post-content").Text()
                words := len(strings.Fields(text))
                readingTime := words / 200 // Assume 200 words per minute
                if readingTime < 1 {
                    readingTime = 1
                }
                return fmt.Sprintf("%d min read", readingTime), nil
            },
            
            "category": func(doc *goquery.Document, url string) (interface{}, error) {
                category := doc.Find(".category, .section, .tag").First().Text()
                return strings.TrimSpace(category), nil
            },
            
            "social_shares": func(doc *goquery.Document, url string) (interface{}, error) {
                shareCount := 0
                doc.Find(".share-count, .social-count").Each(func(i int, s *goquery.Selection) {
                    // Extract share counts if available
                    shareCount++
                })
                return shareCount, nil
            },
            
            "tags": func(doc *goquery.Document, url string) (interface{}, error) {
                var tags []string
                doc.Find(".tags a, .tag, .keyword").Each(func(i int, s *goquery.Selection) {
                    tag := strings.TrimSpace(s.Text())
                    if tag != "" {
                        tags = append(tags, tag)
                    }
                })
                return tags, nil
            },
        },
    }
    
    p := parser.New(opts)
    
    result, err := p.Parse("https://techcrunch.com/2024/01/15/ai-startup-funding/", opts)
    if err != nil {
        log.Fatal("Failed to parse:", err)
    }
    
    if result.IsError() {
        log.Fatal("Extraction failed:", result.Message)
    }
    
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Word Count: %d\n", result.WordCount)
    
    // Access extended fields
    if readingTime, ok := result.Extended["reading_time"].(string); ok {
        fmt.Printf("Reading Time: %s\n", readingTime)
    }
    
    if category, ok := result.Extended["category"].(string); ok && category != "" {
        fmt.Printf("Category: %s\n", category)
    }
    
    if tags, ok := result.Extended["tags"].([]string); ok && len(tags) > 0 {
        fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))
    }
    
    if socialShares, ok := result.Extended["social_shares"].(int); ok {
        fmt.Printf("Social Elements: %d\n", socialShares)
    }
}
```

## Output Format Examples

### Markdown Output

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    opts := &parser.ParserOptions{
        ContentType: "markdown",
    }
    
    p := parser.New(opts)
    
    result, err := p.Parse("https://www.nytimes.com/section/technology", opts)
    if err != nil {
        log.Fatal("Failed to parse:", err)
    }
    
    if result.IsError() {
        log.Fatal("Extraction failed:", result.Message)
    }
    
    // Get formatted markdown with metadata
    markdown := result.FormatMarkdown()
    
    // Save to file
    filename := "article.md"
    err = os.WriteFile(filename, []byte(markdown), 0644)
    if err != nil {
        log.Fatal("Failed to write file:", err)
    }
    
    fmt.Printf("Article saved to %s\n", filename)
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Word Count: %d words\n", result.WordCount)
    
    // Print first few lines of markdown
    lines := strings.Split(markdown, "\n")
    fmt.Println("\nFirst 10 lines of markdown:")
    for i, line := range lines {
        if i >= 10 {
            break
        }
        fmt.Printf("%2d: %s\n", i+1, line)
    }
}
```

### JSON Output

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    p := parser.New()
    
    result, err := p.Parse("https://arstechnica.com/science/", nil)
    if err != nil {
        log.Fatal("Failed to parse:", err)
    }
    
    if result.IsError() {
        log.Fatal("Extraction failed:", result.Message)
    }
    
    // Convert to pretty-printed JSON
    jsonData, err := json.MarshalIndent(result, "", "  ")
    if err != nil {
        log.Fatal("Failed to marshal JSON:", err)
    }
    
    // Save to file
    filename := "article.json"
    err = os.WriteFile(filename, jsonData, 0644)
    if err != nil {
        log.Fatal("Failed to write file:", err)
    }
    
    fmt.Printf("Article data saved to %s\n", filename)
    
    // Print summary
    fmt.Printf("\nExtracted Data Summary:\n")
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Domain: %s\n", result.Domain)
    fmt.Printf("Word Count: %d\n", result.WordCount)
    fmt.Printf("JSON Size: %d bytes\n", len(jsonData))
    
    if result.DatePublished != nil {
        fmt.Printf("Published: %s\n", result.DatePublished.Format("2006-01-02 15:04:05"))
    }
}
```

### Plain Text Output

```go
package main

import (
    "fmt"
    "log"
    "os"
    "strings"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    opts := &parser.ParserOptions{
        ContentType: "text",
    }
    
    p := parser.New(opts)
    
    result, err := p.Parse("https://www.theatlantic.com/technology/", opts)
    if err != nil {
        log.Fatal("Failed to parse:", err)
    }
    
    if result.IsError() {
        log.Fatal("Extraction failed:", result.Message)
    }
    
    // Create formatted text output
    var output strings.Builder
    
    // Add metadata header
    output.WriteString("=" + strings.Repeat("=", len(result.Title)) + "=\n")
    output.WriteString(" " + result.Title + "\n")
    output.WriteString("=" + strings.Repeat("=", len(result.Title)) + "=\n\n")
    
    if result.Author != "" {
        output.WriteString("Author: " + result.Author + "\n")
    }
    
    if result.DatePublished != nil {
        output.WriteString("Published: " + result.DatePublished.Format("January 2, 2006") + "\n")
    }
    
    output.WriteString("Source: " + result.URL + "\n")
    output.WriteString("Word Count: " + fmt.Sprintf("%d", result.WordCount) + " words\n\n")
    
    if result.Dek != "" {
        output.WriteString(result.Dek + "\n\n")
    }
    
    output.WriteString(strings.Repeat("-", 80) + "\n\n")
    
    // Add main content
    output.WriteString(result.Content)
    
    // Save to file
    filename := "article.txt"
    err = os.WriteFile(filename, []byte(output.String()), 0644)
    if err != nil {
        log.Fatal("Failed to write file:", err)
    }
    
    fmt.Printf("Article saved to %s\n", filename)
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Word Count: %d words\n", result.WordCount)
}
```

## Error Handling Examples

### Graceful Error Handling

```go
package main

import (
    "fmt"
    "log"
    "net/url"
    "time"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

type ParseResult struct {
    URL     string
    Result  *parser.Result
    Error   error
    Success bool
}

func parseWithRetry(targetURL string, maxRetries int) *ParseResult {
    p := parser.New()
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        result, err := p.Parse(targetURL, &parser.ParserOptions{
            ContentType: "html",
            Headers: map[string]string{
                "User-Agent": fmt.Sprintf("RetryBot/1.0 (attempt %d)", attempt),
            },
        })
        
        // Check for network errors
        if err != nil {
            log.Printf("Attempt %d failed for %s: %v", attempt, targetURL, err)
            if attempt < maxRetries {
                waitTime := time.Duration(attempt) * time.Second
                log.Printf("Waiting %v before retry...", waitTime)
                time.Sleep(waitTime)
                continue
            }
            return &ParseResult{
                URL:     targetURL,
                Error:   err,
                Success: false,
            }
        }
        
        // Check for extraction errors
        if result.IsError() {
            log.Printf("Extraction failed for %s: %s", targetURL, result.Message)
            return &ParseResult{
                URL:     targetURL,
                Result:  result,
                Error:   fmt.Errorf("extraction failed: %s", result.Message),
                Success: false,
            }
        }
        
        // Validate content quality
        if result.Title == "" {
            log.Printf("No title extracted for %s", targetURL)
        }
        
        if result.WordCount < 100 {
            log.Printf("Warning: Short content for %s (%d words)", targetURL, result.WordCount)
        }
        
        return &ParseResult{
            URL:     targetURL,
            Result:  result,
            Success: true,
        }
    }
    
    return &ParseResult{
        URL:     targetURL,
        Error:   fmt.Errorf("failed after %d attempts", maxRetries),
        Success: false,
    }
}

func main() {
    urls := []string{
        "https://www.example.com/valid-article",
        "https://invalid-domain-that-does-not-exist.com",
        "https://httpbin.org/status/404",
        "https://www.theguardian.com/technology",
    }
    
    for _, url := range urls {
        fmt.Printf("\nProcessing: %s\n", url)
        
        // Parse URL first
        if _, err := url.Parse(url); err != nil {
            log.Printf("Invalid URL format: %v", err)
            continue
        }
        
        result := parseWithRetry(url, 3)
        
        if result.Success {
            fmt.Printf("✓ Success: %s\n", result.Result.Title)
            fmt.Printf("  Author: %s\n", result.Result.Author)
            fmt.Printf("  Word Count: %d\n", result.Result.WordCount)
        } else {
            fmt.Printf("✗ Failed: %v\n", result.Error)
        }
    }
}
```

### Error Classification

```go
package main

import (
    "fmt"
    "log"
    "net"
    "net/url"
    "strings"
    "time"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

type ErrorCategory int

const (
    NetworkError ErrorCategory = iota
    ValidationError
    ExtractionError
    ContentError
    UnknownError
)

func (e ErrorCategory) String() string {
    switch e {
    case NetworkError:
        return "Network Error"
    case ValidationError:
        return "Validation Error"
    case ExtractionError:
        return "Extraction Error"
    case ContentError:
        return "Content Error"
    default:
        return "Unknown Error"
    }
}

type CategorizedError struct {
    Category    ErrorCategory
    Message     string
    URL         string
    Recoverable bool
    RetryAfter  time.Duration
}

func classifyError(targetURL string, err error, result *parser.Result) *CategorizedError {
    // Network errors
    if err != nil {
        if netErr, ok := err.(net.Error); ok {
            if netErr.Timeout() {
                return &CategorizedError{
                    Category:    NetworkError,
                    Message:     "Request timeout",
                    URL:         targetURL,
                    Recoverable: true,
                    RetryAfter:  5 * time.Second,
                }
            }
        }
        
        if strings.Contains(err.Error(), "connection refused") {
            return &CategorizedError{
                Category:    NetworkError,
                Message:     "Connection refused",
                URL:         targetURL,
                Recoverable: true,
                RetryAfter:  10 * time.Second,
            }
        }
        
        if strings.Contains(err.Error(), "no such host") {
            return &CategorizedError{
                Category:    NetworkError,
                Message:     "Domain not found",
                URL:         targetURL,
                Recoverable: false,
            }
        }
        
        return &CategorizedError{
            Category:    NetworkError,
            Message:     err.Error(),
            URL:         targetURL,
            Recoverable: true,
            RetryAfter:  3 * time.Second,
        }
    }
    
    // Extraction errors
    if result != nil && result.IsError() {
        return &CategorizedError{
            Category:    ExtractionError,
            Message:     result.Message,
            URL:         targetURL,
            Recoverable: false,
        }
    }
    
    // Content validation errors
    if result != nil {
        if result.Title == "" && result.Content == "" {
            return &CategorizedError{
                Category:    ContentError,
                Message:     "No content extracted",
                URL:         targetURL,
                Recoverable: false,
            }
        }
        
        if result.WordCount < 50 {
            return &CategorizedError{
                Category:    ContentError,
                Message:     fmt.Sprintf("Content too short: %d words", result.WordCount),
                URL:         targetURL,
                Recoverable: false,
            }
        }
    }
    
    return nil // No error
}

func parseWithErrorHandling(targetURL string) {
    // Validate URL format first
    if _, err := url.Parse(targetURL); err != nil {
        log.Printf("❌ %s: Invalid URL format: %v", targetURL, err)
        return
    }
    
    p := parser.New()
    
    start := time.Now()
    result, err := p.Parse(targetURL, nil)
    duration := time.Since(start)
    
    // Classify any errors
    if categorizedErr := classifyError(targetURL, err, result); categorizedErr != nil {
        fmt.Printf("❌ %s [%s]: %s\n", 
            categorizedErr.URL, 
            categorizedErr.Category, 
            categorizedErr.Message)
        
        if categorizedErr.Recoverable {
            fmt.Printf("   → Retry recommended after %v\n", categorizedErr.RetryAfter)
        } else {
            fmt.Printf("   → Not recoverable\n")
        }
        return
    }
    
    // Success
    fmt.Printf("✅ %s: %s\n", targetURL, result.Title)
    fmt.Printf("   Author: %s | Words: %d | Time: %v\n", 
        result.Author, result.WordCount, duration)
    
    // Content quality warnings
    if result.Author == "" {
        fmt.Printf("   ⚠️  No author found\n")
    }
    
    if result.DatePublished == nil {
        fmt.Printf("   ⚠️  No publish date found\n")
    }
    
    if result.WordCount < 200 {
        fmt.Printf("   ⚠️  Short article (%d words)\n", result.WordCount)
    }
}

func main() {
    testURLs := []string{
        "https://www.nytimes.com/section/technology",        // Should work
        "https://invalid-domain-12345.com/article",          // Domain not found
        "https://httpbin.org/delay/10",                      // Timeout
        "https://httpbin.org/status/500",                    // Server error
        "https://www.google.com",                            // No article content
        "not-a-valid-url",                                   // Invalid URL format
    }
    
    fmt.Println("Testing error handling with various scenarios:")
    fmt.Println(strings.Repeat("=", 60))
    
    for _, url := range testURLs {
        parseWithErrorHandling(url)
        fmt.Println()
    }
}
```

## Batch Processing Examples

### Simple Batch Processing

```go
package main

import (
    "fmt"
    "log"
    "sync"
    "time"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func processBatch(urls []string) {
    // Use high-throughput parser for better performance
    htParser := parser.NewHighThroughputParser(&parser.ParserOptions{
        ContentType: "markdown",
        Headers: map[string]string{
            "User-Agent": "BatchProcessor/1.0",
        },
    })
    
    fmt.Printf("Processing %d URLs...\n", len(urls))
    start := time.Now()
    
    // Process all URLs in parallel
    results, err := htParser.BatchParse(urls, nil)
    if err != nil {
        log.Fatal("Batch processing failed:", err)
    }
    
    totalDuration := time.Since(start)
    
    // Analyze results
    successful := 0
    totalWords := 0
    
    for i, result := range results {
        if result.IsError() {
            fmt.Printf("❌ URL %d: %s\n", i+1, result.Message)
        } else {
            successful++
            totalWords += result.WordCount
            fmt.Printf("✅ URL %d: %s (%d words)\n", i+1, result.Title, result.WordCount)
        }
        
        // Return result to pool for memory efficiency
        htParser.ReturnResult(result)
    }
    
    // Print statistics
    fmt.Printf("\nBatch Processing Summary:\n")
    fmt.Printf("Total URLs: %d\n", len(urls))
    fmt.Printf("Successful: %d\n", successful)
    fmt.Printf("Failed: %d\n", len(urls)-successful)
    fmt.Printf("Total words extracted: %d\n", totalWords)
    fmt.Printf("Total time: %v\n", totalDuration)
    fmt.Printf("Average time per URL: %v\n", totalDuration/time.Duration(len(urls)))
    
    if successful > 0 {
        fmt.Printf("Average words per article: %d\n", totalWords/successful)
    }
    
    // Print performance stats
    stats := htParser.GetStats()
    fmt.Printf("Pool efficiency: %.1f%%\n", stats.PoolEfficiency()*100)
}

func main() {
    urls := []string{
        "https://www.theguardian.com/technology/ai",
        "https://www.nytimes.com/section/technology",
        "https://techcrunch.com/category/artificial-intelligence/",
        "https://arstechnica.com/science/",
        "https://www.wired.com/category/science/",
        "https://www.theatlantic.com/technology/",
        "https://www.reuters.com/technology/",
        "https://www.bbc.com/news/technology",
    }
    
    processBatch(urls)
}
```

### Concurrent Processing with Rate Limiting

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"
    
    "golang.org/x/time/rate"
    "github.com/BumpyClock/hermes/pkg/parser"
)

type Result struct {
    URL      string
    Title    string
    Words    int
    Duration time.Duration
    Error    error
}

func processURLsWithRateLimit(urls []string, maxConcurrency int, requestsPerSecond int) []Result {
    // Create rate limiter
    limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), 1)
    
    // Create semaphore for concurrency control
    semaphore := make(chan struct{}, maxConcurrency)
    
    // Results collection
    results := make([]Result, len(urls))
    var wg sync.WaitGroup
    
    p := parser.New(&parser.ParserOptions{
        ContentType: "html",
        Headers: map[string]string{
            "User-Agent": "RateLimitedBot/1.0",
        },
    })
    
    fmt.Printf("Processing %d URLs with %d max concurrency and %d req/sec rate limit\n", 
        len(urls), maxConcurrency, requestsPerSecond)
    
    for i, url := range urls {
        wg.Add(1)
        
        go func(index int, targetURL string) {
            defer wg.Done()
            
            // Acquire semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            // Wait for rate limiter
            err := limiter.Wait(context.Background())
            if err != nil {
                results[index] = Result{
                    URL:   targetURL,
                    Error: fmt.Errorf("rate limit error: %w", err),
                }
                return
            }
            
            // Process URL
            start := time.Now()
            result, err := p.Parse(targetURL, nil)
            duration := time.Since(start)
            
            if err != nil {
                results[index] = Result{
                    URL:      targetURL,
                    Duration: duration,
                    Error:    err,
                }
                return
            }
            
            if result.IsError() {
                results[index] = Result{
                    URL:      targetURL,
                    Duration: duration,
                    Error:    fmt.Errorf("extraction failed: %s", result.Message),
                }
                return
            }
            
            results[index] = Result{
                URL:      targetURL,
                Title:    result.Title,
                Words:    result.WordCount,
                Duration: duration,
            }
            
            fmt.Printf("✅ Processed %s (%v)\n", targetURL, duration)
            
        }(i, url)
    }
    
    wg.Wait()
    return results
}

func main() {
    urls := []string{
        "https://www.example.com/article1",
        "https://www.example.com/article2",
        "https://www.example.com/article3",
        "https://www.example.com/article4",
        "https://www.example.com/article5",
        "https://httpbin.org/html",  // Test URL that works
        "https://httpbin.org/delay/1",
        "https://httpbin.org/delay/2",
    }
    
    start := time.Now()
    results := processURLsWithRateLimit(urls, 3, 2) // 3 concurrent, 2 req/sec
    totalDuration := time.Since(start)
    
    // Analyze results
    successful := 0
    failed := 0
    totalWords := 0
    totalProcessingTime := time.Duration(0)
    
    fmt.Printf("\nResults:\n")
    fmt.Printf(strings.Repeat("=", 80) + "\n")
    
    for i, result := range results {
        if result.Error != nil {
            fmt.Printf("%d. ❌ %s\n   Error: %v\n", i+1, result.URL, result.Error)
            failed++
        } else {
            fmt.Printf("%d. ✅ %s\n   Title: %s\n   Words: %d\n   Time: %v\n", 
                i+1, result.URL, result.Title, result.Words, result.Duration)
            successful++
            totalWords += result.Words
            totalProcessingTime += result.Duration
        }
        fmt.Println()
    }
    
    fmt.Printf("Summary:\n")
    fmt.Printf("Total URLs: %d\n", len(urls))
    fmt.Printf("Successful: %d\n", successful)
    fmt.Printf("Failed: %d\n", failed)
    fmt.Printf("Wall clock time: %v\n", totalDuration)
    fmt.Printf("Total processing time: %v\n", totalProcessingTime)
    fmt.Printf("Concurrency efficiency: %.1f%%\n", 
        float64(totalProcessingTime)/float64(totalDuration)*100)
    
    if successful > 0 {
        fmt.Printf("Average words per article: %d\n", totalWords/successful)
        fmt.Printf("Average processing time: %v\n", totalProcessingTime/time.Duration(successful))
    }
}
```

## Real-World Use Cases

### RSS Feed Processing

```go
package main

import (
    "encoding/xml"
    "fmt"
    "log"
    "net/http"
    "strings"
    "time"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

type RSSFeed struct {
    Title string `xml:"channel>title"`
    Items []RSSItem `xml:"channel>item"`
}

type RSSItem struct {
    Title   string `xml:"title"`
    Link    string `xml:"link"`
    PubDate string `xml:"pubDate"`
}

type Article struct {
    Title       string
    URL         string
    Author      string
    Content     string
    WordCount   int
    PublishDate time.Time
    Excerpt     string
}

func fetchRSSFeed(feedURL string) (*RSSFeed, error) {
    resp, err := http.Get(feedURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var feed RSSFeed
    err = xml.NewDecoder(resp.Body).Decode(&feed)
    if err != nil {
        return nil, err
    }
    
    return &feed, nil
}

func processRSSFeed(feedURL string, maxArticles int) ([]Article, error) {
    // Fetch RSS feed
    feed, err := fetchRSSFeed(feedURL)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
    }
    
    fmt.Printf("Processing RSS feed: %s\n", feed.Title)
    fmt.Printf("Found %d items\n", len(feed.Items))
    
    // Limit number of articles to process
    items := feed.Items
    if len(items) > maxArticles {
        items = items[:maxArticles]
    }
    
    // Extract URLs
    urls := make([]string, len(items))
    for i, item := range items {
        urls[i] = item.Link
    }
    
    // Use high-throughput parser for batch processing
    htParser := parser.NewHighThroughputParser(&parser.ParserOptions{
        ContentType: "text",
        Headers: map[string]string{
            "User-Agent": "RSSProcessor/1.0 (+https://example.com/bot)",
        },
    })
    
    // Process all articles
    results, err := htParser.BatchParse(urls, nil)
    if err != nil {
        return nil, fmt.Errorf("batch processing failed: %w", err)
    }
    
    // Convert to articles
    articles := make([]Article, 0, len(results))
    
    for i, result := range results {
        if result.IsError() {
            log.Printf("Failed to process %s: %s", urls[i], result.Message)
            htParser.ReturnResult(result)
            continue
        }
        
        article := Article{
            Title:     result.Title,
            URL:       result.URL,
            Author:    result.Author,
            Content:   result.Content,
            WordCount: result.WordCount,
            Excerpt:   result.Excerpt,
        }
        
        if result.DatePublished != nil {
            article.PublishDate = *result.DatePublished
        }
        
        articles = append(articles, article)
        htParser.ReturnResult(result)
    }
    
    return articles, nil
}

func main() {
    feedURLs := []string{
        "https://feeds.arstechnica.com/arstechnica/index",
        "https://www.wired.com/feed/rss",
        "https://techcrunch.com/feed/",
    }
    
    for _, feedURL := range feedURLs {
        fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
        fmt.Printf("Processing feed: %s\n", feedURL)
        fmt.Printf(strings.Repeat("=", 80) + "\n")
        
        articles, err := processRSSFeed(feedURL, 5) // Process 5 articles per feed
        if err != nil {
            log.Printf("Failed to process feed %s: %v", feedURL, err)
            continue
        }
        
        fmt.Printf("\nSuccessfully processed %d articles:\n\n", len(articles))
        
        totalWords := 0
        for i, article := range articles {
            fmt.Printf("%d. %s\n", i+1, article.Title)
            fmt.Printf("   Author: %s\n", article.Author)
            fmt.Printf("   Words: %d\n", article.WordCount)
            fmt.Printf("   URL: %s\n", article.URL)
            
            if !article.PublishDate.IsZero() {
                fmt.Printf("   Published: %s\n", article.PublishDate.Format("Jan 2, 2006"))
            }
            
            if article.Excerpt != "" {
                excerpt := article.Excerpt
                if len(excerpt) > 100 {
                    excerpt = excerpt[:100] + "..."
                }
                fmt.Printf("   Excerpt: %s\n", excerpt)
            }
            
            fmt.Println()
            totalWords += article.WordCount
        }
        
        if len(articles) > 0 {
            avgWords := totalWords / len(articles)
            fmt.Printf("Average article length: %d words\n", avgWords)
        }
    }
}
```

This comprehensive set of examples demonstrates the versatility and power of Hermes for various content extraction scenarios, from simple single-page extraction to complex batch processing workflows.