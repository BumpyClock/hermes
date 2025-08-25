# Basic Usage Guide

This guide covers fundamental usage patterns for Hermes, from simple content extraction to advanced configuration options.

## Table of Contents

- [Command Line Usage](#command-line-usage)
- [Library Usage](#library-usage)
- [Output Formats](#output-formats)
- [Common Patterns](#common-patterns)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Command Line Usage

The Hermes CLI provides a simple interface for extracting content from web pages.

### Basic Commands

#### Parse a Single URL

```bash
# Basic parsing with JSON output
parser parse https://example.com/article

# Parse with specific output format
parser parse -f markdown https://example.com/article
parser parse -f html https://example.com/article
parser parse -f text https://example.com/article
```

#### Save Output to File

```bash
# Save as Markdown
parser parse -f markdown -o article.md https://example.com/article

# Save as HTML
parser parse -f html -o article.html https://example.com/article

# Save as JSON
parser parse -o article.json https://example.com/article
```

#### Parse Multiple URLs

```bash
# Parse multiple URLs (outputs JSON array)
parser parse https://example.com/1 https://example.com/2 https://example.com/3

# With timing information
parser parse --timing https://example.com/1 https://example.com/2

# Save multiple results
parser parse -o articles.json https://example.com/1 https://example.com/2
```

### Advanced CLI Options

#### Custom Headers

```bash
# Single header
parser parse --headers '{"User-Agent": "MyBot/1.0"}' https://example.com

# Multiple headers
parser parse --headers '{
  "User-Agent": "MyBot/1.0",
  "Accept": "text/html,application/xhtml+xml",
  "Accept-Language": "en-US,en;q=0.5"
}' https://example.com
```

#### Control Pagination

```bash
# Disable pagination handling hint (default is true)
parser parse --fetch-all=false https://example.com/article

# Enable pagination handling hint (default)
parser parse --fetch-all=true https://example.com/article
```

**Note on Multi-page Articles:** Hermes detects and exposes `next_page_url` when a site provides a "next page" link. However, automatic fetching and merging of subsequent pages is not yet implemented. Use the `next_page_url` value to iterate manually if needed.

#### Performance Monitoring

```bash
# Show timing for each URL
parser parse --timing https://example.com/1 https://example.com/2

# Example output:
# Parsing URL 1/2: https://example.com/1
# Parsed https://example.com/1 in 1.2s
# Parsing URL 2/2: https://example.com/2
# Parsed https://example.com/2 in 0.8s
#
# Timing Summary:
# Total URLs: 2
# Total parse time: 2s
# Average parse time: 1s
```

### CLI Examples

#### News Article Extraction

```bash
# Extract NY Times article as Markdown
parser parse -f markdown -o nytimes-article.md \
  "https://www.nytimes.com/2024/01/15/technology/ai-breakthrough.html"

# Extract with custom User-Agent
parser parse --headers '{"User-Agent": "NewsBot/1.0"}' \
  "https://www.theguardian.com/technology/ai"
```

#### Batch Processing

```bash
# Process multiple tech articles
parser parse --timing -f markdown -o tech-articles.json \
  "https://arstechnica.com/latest-article" \
  "https://www.theverge.com/tech-news" \
  "https://techcrunch.com/startup-news"
```

#### Blog Post Extraction

```bash
# Extract blog post with full content
parser parse -f html --fetch-all=true \
  "https://medium.com/@author/long-article-part-1"
```

## Library Usage

Use Hermes as a Go library for programmatic content extraction.

### Basic Library Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    // Create parser with default options
    p := parser.New()
    
    // Parse a URL
    result, err := p.Parse("https://example.com/article", nil)
    if err != nil {
        log.Fatal("Failed to parse:", err)
    }
    
    // Check for extraction errors
    if result.IsError() {
        log.Fatal("Extraction failed:", result.Message)
    }
    
    // Use the extracted content
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Word Count: %d\n", result.WordCount)
    
    // Site metadata is automatically extracted
    if result.Language != "" {
        fmt.Printf("Language: %s\n", result.Language)
    }
    if result.Description != "" {
        fmt.Printf("Description: %s\n", result.Description)
    }
    
    fmt.Printf("Content: %s\n", result.Content)
}
```

### Advanced Library Usage

#### Custom Configuration

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    // Create parser with custom options
    opts := &parser.ParserOptions{
        ContentType:   "markdown",
        FetchAllPages: true, // Enables next_page_url detection only
        Fallback:      true,
        Headers: map[string]string{
            "User-Agent": "MyApp/1.0",
            "Accept":     "text/html,application/xhtml+xml",
        },
    }
    
    p := parser.New(opts)
    
    // Parse with options
    result, err := p.Parse("https://example.com/article", opts)
    if err != nil {
        log.Fatal(err)
    }
    
    if result.IsError() {
        log.Fatal(result.Message)
    }
    
    // Get Markdown content
    markdown := result.FormatMarkdown()
    fmt.Println(markdown)
}
```

#### Parsing HTML Directly

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    htmlContent := `
    <html>
        <head><title>Sample Article</title></head>
        <body>
            <h1>Article Title</h1>
            <p class="byline">By John Doe</p>
            <div class="content">
                <p>This is the article content...</p>
                <p>More content here...</p>
            </div>
        </body>
    </html>`
    
    p := parser.New()
    
    // Parse HTML directly
    result, err := p.ParseHTML(htmlContent, "https://example.com", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Extracted Title: %s\n", result.Title)
    fmt.Printf("Extracted Content: %s\n", result.Content)
}
```

### High-Performance Usage

#### Batch Processing

```go
package main

import (
    "fmt"
    "log"
    "sync"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    urls := []string{
        "https://example.com/article1",
        "https://example.com/article2", 
        "https://example.com/article3",
    }
    
    // Use high-throughput parser for batch processing
    htParser := parser.NewHighThroughputParser(&parser.ParserOptions{
        ContentType: "markdown",
    })
    
    // Process all URLs in parallel
    results, err := htParser.BatchParse(urls, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process results
    for i, result := range results {
        if result.IsError() {
            fmt.Printf("URL %d failed: %s\n", i+1, result.Message)
            continue
        }
        
        fmt.Printf("URL %d: %s (%d words)\n", 
            i+1, result.Title, result.WordCount)
            
        // Return result to pool for memory efficiency
        htParser.ReturnResult(result)
    }
    
    // Print performance stats
    stats := htParser.GetStats()
    fmt.Printf("Performance: %d%% pool efficiency\n", 
        int(stats.PoolEfficiency()*100))
}
```

#### Concurrent Processing

```go
package main

import (
    "fmt"
    "log"
    "sync"
    
    "github.com/BumpyClock/hermes/pkg/parser"
)

func processURL(url string, p *parser.Hermes, wg *sync.WaitGroup, results chan<- *parser.Result) {
    defer wg.Done()
    
    result, err := p.Parse(url, nil)
    if err != nil {
        log.Printf("Failed to parse %s: %v", url, err)
        return
    }
    
    results <- result
}

func main() {
    urls := []string{
        "https://example.com/1",
        "https://example.com/2",
        "https://example.com/3",
    }
    
    p := parser.New()
    results := make(chan *parser.Result, len(urls))
    var wg sync.WaitGroup
    
    // Start concurrent processing
    for _, url := range urls {
        wg.Add(1)
        go processURL(url, p, &wg, results)
    }
    
    // Wait for completion
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    for result := range results {
        if !result.IsError() {
            fmt.Printf("Title: %s\n", result.Title)
        }
    }
}
```

## Output Formats

Hermes supports multiple output formats for different use cases.

### HTML Output

Clean, semantic HTML suitable for web display:

```go
opts := &parser.ParserOptions{ContentType: "html"}
result, _ := parser.Parse(url, opts)

fmt.Println(result.Content)
// Output:
// <h2>Section Title</h2>
// <p>Article content with <strong>emphasis</strong>.</p>
// <ul><li>List item</li></ul>
```

### Markdown Output

Clean Markdown for documentation and text processing:

```go
opts := &parser.ParserOptions{ContentType: "markdown"}
result, _ := parser.Parse(url, opts)

// Get formatted markdown with metadata
markdown := result.FormatMarkdown()
fmt.Println(markdown)

// Output:
// # Article Title
// 
// **Author:** John Doe
// **Date:** 2024-01-15T10:30:00Z
// **URL:** https://example.com/article
// 
// ## Section Title
// 
// Article content with **emphasis**.
// 
// - List item
```

### Text Output

Plain text with minimal formatting:

```go
opts := &parser.ParserOptions{ContentType: "text"}
result, _ := parser.Parse(url, opts)

fmt.Println(result.Content)
// Output:
// Section Title
// 
// Article content with emphasis.
// 
// - List item
```

### JSON Output

Complete structured data:

```go
import "encoding/json"

result, _ := parser.Parse(url, nil)
jsonData, _ := json.MarshalIndent(result, "", "  ")
fmt.Println(string(jsonData))

// Output:
// {
//   "title": "Article Title",
//   "content": "<p>Article content...</p>",
//   "author": "John Doe",
//   "date_published": "2024-01-15T10:30:00Z",
//   "word_count": 450,
//   "url": "https://example.com/article",
//   "language": "en-US",
//   "description": "Site description from meta tags"
// }
```

## Common Patterns

### Pattern 1: News Article Processing

```go
func processNewsArticle(url string) (*parser.Result, error) {
    opts := &parser.ParserOptions{
        ContentType:   "markdown",
        FetchAllPages: true,
        Headers: map[string]string{
            "User-Agent": "NewsBot/1.0 (+https://yoursite.com/bot)",
        },
    }
    
    p := parser.New(opts)
    result, err := p.Parse(url, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch article: %w", err)
    }
    
    if result.IsError() {
        return nil, fmt.Errorf("failed to extract content: %s", result.Message)
    }
    
    // Validate minimum content requirements
    if result.Title == "" {
        return nil, errors.New("no title found")
    }
    
    if result.WordCount < 100 {
        return nil, errors.New("article too short")
    }
    
    return result, nil
}
```

### Pattern 2: Blog Post Archiving

```go
func archiveBlogPost(url, outputDir string) error {
    opts := &parser.ParserOptions{
        ContentType:   "markdown",
        FetchAllPages: true,
    }
    
    p := parser.New(opts)
    result, err := p.Parse(url, opts)
    if err != nil {
        return err
    }
    
    if result.IsError() {
        return errors.New(result.Message)
    }
    
    // Create filename from title
    filename := strings.ReplaceAll(result.Title, " ", "-")
    filename = strings.ToLower(filename) + ".md"
    filepath := filepath.Join(outputDir, filename)
    
    // Write markdown with metadata
    content := result.FormatMarkdown()
    return os.WriteFile(filepath, []byte(content), 0644)
}
```

### Pattern 3: Content Analysis

```go
type ContentStats struct {
    Title       string
    Author      string
    WordCount   int
    ReadingTime int // minutes
    Domain      string
    PublishDate time.Time
    Category    string
    Language    string // Content language
    Description string // Site description
}

func analyzeContent(url string) (*ContentStats, error) {
    opts := &parser.ParserOptions{
        Extend: map[string]parser.ExtractorFunc{
            "category": func(doc *goquery.Document, url string) (interface{}, error) {
                return doc.Find(".category, .tag, .section").First().Text(), nil
            },
        },
    }
    
    p := parser.New(opts)
    result, err := p.Parse(url, opts)
    if err != nil {
        return nil, err
    }
    
    if result.IsError() {
        return nil, errors.New(result.Message)
    }
    
    stats := &ContentStats{
        Title:       result.Title,
        Author:      result.Author,
        WordCount:   result.WordCount,
        ReadingTime: result.WordCount / 200, // 200 WPM average
        Domain:      result.Domain,
        Language:    result.Language,
        Description: result.Description,
    }
    
    if result.DatePublished != nil {
        stats.PublishDate = *result.DatePublished
    }
    
    if category, ok := result.Extended["category"].(string); ok {
        stats.Category = category
    }
    
    return stats, nil
}
```

### Pattern 4: RSS Feed Processing

```go
func processFeedUrls(urls []string) ([]*parser.Result, error) {
    opts := &parser.ParserOptions{
        ContentType: "html",
        Fallback:    true,
    }
    
    htParser := parser.NewHighThroughputParser(opts)
    
    results, err := htParser.BatchParse(urls, opts)
    if err != nil {
        return nil, err
    }
    
    // Filter successful results
    successful := make([]*parser.Result, 0, len(results))
    for _, result := range results {
        if !result.IsError() && result.WordCount > 100 {
            successful = append(successful, result)
        } else {
            // Return failed results to pool
            htParser.ReturnResult(result)
        }
    }
    
    return successful, nil
}
```

## Error Handling

### Graceful Error Handling

```go
func parseWithRetry(url string, maxRetries int) (*parser.Result, error) {
    p := parser.New()
    
    for attempt := 1; attempt <= maxRetries; attempt++ {
        result, err := p.Parse(url, nil)
        
        // Network error - retry
        if err != nil {
            if attempt < maxRetries {
                time.Sleep(time.Duration(attempt) * time.Second)
                continue
            }
            return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, err)
        }
        
        // Extraction error - don't retry
        if result.IsError() {
            return nil, fmt.Errorf("extraction failed: %s", result.Message)
        }
        
        return result, nil
    }
    
    return nil, fmt.Errorf("unexpected error after %d attempts", maxRetries)
}
```

### Error Classification

```go
type ErrorType int

const (
    NetworkError ErrorType = iota
    ExtractionError
    ValidationError
)

type ParseError struct {
    Type    ErrorType
    URL     string
    Message string
    Err     error
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("%s: %s", e.URL, e.Message)
}

func classifyError(url string, err error, result *parser.Result) *ParseError {
    if err != nil {
        return &ParseError{
            Type:    NetworkError,
            URL:     url,
            Message: "Failed to fetch content",
            Err:     err,
        }
    }
    
    if result.IsError() {
        return &ParseError{
            Type:    ExtractionError,
            URL:     url,
            Message: result.Message,
        }
    }
    
    if result.Title == "" || result.WordCount < 50 {
        return &ParseError{
            Type:    ValidationError,
            URL:     url,
            Message: "Insufficient content extracted",
        }
    }
    
    return nil
}
```

## Best Practices

### 1. Resource Management

```go
// Use high-throughput parser for batch operations
htParser := parser.NewHighThroughputParser(opts)

// Always return results to pool
defer htParser.ReturnResult(result)

// Monitor performance
stats := htParser.GetStats()
if stats.PoolEfficiency() < 0.8 {
    log.Printf("Poor pool efficiency: %.2f%%", stats.PoolEfficiency()*100)
}
```

### 2. Rate Limiting

```go
import "golang.org/x/time/rate"

func parseWithRateLimit(urls []string) {
    // Limit to 10 requests per second
    limiter := rate.NewLimiter(10, 1)
    
    p := parser.New()
    
    for _, url := range urls {
        // Wait for rate limit
        limiter.Wait(context.Background())
        
        result, err := p.Parse(url, nil)
        if err != nil {
            log.Printf("Failed to parse %s: %v", url, err)
            continue
        }
        
        // Process result...
    }
}
```

### 3. Timeout Handling

```go
func parseWithTimeout(url string, timeout time.Duration) (*parser.Result, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    // Use context for timeout control
    opts := &parser.ParserOptions{
        Headers: map[string]string{
            "User-Agent": "TimeoutBot/1.0",
        },
    }
    
    // Create channel for result
    resultChan := make(chan *parser.Result, 1)
    errorChan := make(chan error, 1)
    
    // Parse in goroutine
    go func() {
        p := parser.New(opts)
        result, err := p.Parse(url, opts)
        if err != nil {
            errorChan <- err
            return
        }
        resultChan <- result
    }()
    
    // Wait for result or timeout
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errorChan:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### 4. Content Validation

```go
func validateContent(result *parser.Result) error {
    if result.IsError() {
        return errors.New(result.Message)
    }
    
    if result.Title == "" {
        return errors.New("no title extracted")
    }
    
    if result.WordCount < 100 {
        return errors.New("content too short")
    }
    
    if result.Author == "" {
        log.Printf("Warning: no author found for %s", result.URL)
    }
    
    if result.DatePublished == nil {
        log.Printf("Warning: no publish date found for %s", result.URL)
    }
    
    return nil
}
```

### 5. Logging and Monitoring

```go
import "go.uber.org/zap"

func parseWithLogging(url string) (*parser.Result, error) {
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    start := time.Now()
    
    p := parser.New()
    result, err := p.Parse(url, nil)
    
    duration := time.Since(start)
    
    if err != nil {
        logger.Error("Parse failed",
            zap.String("url", url),
            zap.Error(err),
            zap.Duration("duration", duration),
        )
        return nil, err
    }
    
    if result.IsError() {
        logger.Warn("Extraction failed",
            zap.String("url", url),
            zap.String("error", result.Message),
            zap.Duration("duration", duration),
        )
        return nil, errors.New(result.Message)
    }
    
    logger.Info("Parse successful",
        zap.String("url", url),
        zap.String("title", result.Title),
        zap.Int("word_count", result.WordCount),
        zap.Duration("duration", duration),
    )
    
    return result, nil
}
```

## Next Steps

After mastering basic usage:

1. Learn about [Custom Extractors](custom-extractors.md)
2. Explore [Advanced Configuration](advanced-config.md)
3. Check [Integration Examples](integration.md)
4. Review [Performance Optimization](../architecture/performance.md)
