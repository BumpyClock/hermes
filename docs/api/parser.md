# Parser API Reference

The `parser` package provides the core functionality for extracting clean, structured content from web pages.

## Table of Contents

- [Types](#types)
- [Core Functions](#core-functions)
- [Parser Interface](#parser-interface)
- [Mercury Parser](#mercury-parser)
- [High Throughput Parser](#high-throughput-parser)
- [Batch Operations](#batch-operations)
- [Error Handling](#error-handling)

## Types

### Parser

```go
type Parser interface {
    Parse(url string, opts *ParserOptions) (*Result, error)
    ParseHTML(html string, url string, opts *ParserOptions) (*Result, error)
}
```

Main interface for content extraction operations.

### Mercury

```go
type Mercury struct {
    options  ParserOptions
    htParser *HighThroughputParser
}
```

Main parser implementation with built-in optimizations and pooling.

### ParserOptions

```go
type ParserOptions struct {
    FetchAllPages   bool              // Enable next page URL detection (merging not implemented)
    Fallback        bool              // Use generic extractor as fallback
    ContentType     string            // Output format: "html", "markdown", "text"
    Headers         map[string]string // Custom HTTP headers
    CustomExtractor *CustomExtractor  // Custom extraction rules
    Extend          map[string]ExtractorFunc // Extended fields
}
```

Configuration options for parser behavior.

#### Fields

- **FetchAllPages** (bool): Enable next page URL detection (automatic merging not yet implemented)
- **Fallback** (bool): Fall back to generic extractor if custom extractor fails
- **ContentType** (string): Output format - "html", "markdown", or "text"
- **Headers** (map[string]string): Custom HTTP headers for requests
- **CustomExtractor** (*CustomExtractor): Override site-specific extraction rules
- **Extend** (map[string]ExtractorFunc): Add custom field extractors

### Result

```go
type Result struct {
    Title          string                 `json:"title"`
    Content        string                 `json:"content"`
    Author         string                 `json:"author"`
    DatePublished  *time.Time            `json:"date_published"`
    LeadImageURL   string                `json:"lead_image_url"`
    Dek            string                `json:"dek"`
    NextPageURL    string                `json:"next_page_url"`
    URL            string                `json:"url"`
    Domain         string                `json:"domain"`
    Excerpt        string                `json:"excerpt"`
    WordCount      int                   `json:"word_count"`
    Direction      string                `json:"direction"`
    TotalPages     int                   `json:"total_pages"`
    RenderedPages  int                   `json:"rendered_pages"`
    ExtractorUsed  string                `json:"extractor_used,omitempty"`
    Extended       map[string]interface{} `json:"extended,omitempty"`
    
    // Site metadata fields
    Description    string                `json:"description"`
    Language       string                `json:"language"`
    
    Error          bool                   `json:"error,omitempty"`
    Message        string                 `json:"message,omitempty"`
}
```

Extracted article data and metadata.

## Core Functions

### New

```go
func New(opts ...*ParserOptions) *Mercury
```

Creates a new optimized Mercury parser instance.

**Parameters:**
- `opts` (optional): Parser configuration options

**Returns:**
- `*Mercury`: Configured parser instance

**Example:**
```go
// Default configuration
parser := parser.New()

// Custom configuration
opts := &parser.ParserOptions{
    ContentType: "markdown",
    FetchAllPages: true,
}
parser := parser.New(opts)
```

### NewParser

```go
func NewParser() *Mercury
```

Convenience function to create a new parser with default options.

**Returns:**
- `*Mercury`: Parser instance with defaults

### DefaultParserOptions

```go
func DefaultParserOptions() *ParserOptions
```

Returns default parser configuration.

**Returns:**
- `*ParserOptions`: Default configuration

**Default Values:**
```go
&ParserOptions{
    FetchAllPages: true,
    Fallback:      true,
    ContentType:   "html",
}
```

## Parser Interface

### Parse

```go
func (m *Mercury) Parse(targetURL string, opts *ParserOptions) (*Result, error)
```

Extracts content from a URL using optimized pooling.

**Parameters:**
- `targetURL` (string): URL to extract content from
- `opts` (*ParserOptions): Optional configuration (uses parser defaults if nil)

**Returns:**
- `*Result`: Extracted content and metadata
- `error`: Extraction or network error

**Example:**
```go
result, err := parser.Parse("https://example.com/article", &parser.ParserOptions{
    ContentType: "markdown",
    Headers: map[string]string{
        "User-Agent": "MyBot/1.0",
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Title: %s\n", result.Title)
fmt.Printf("Author: %s\n", result.Author)
fmt.Printf("Content: %s\n", result.Content)
```

### ParseHTML

```go
func (m *Mercury) ParseHTML(html string, targetURL string, opts *ParserOptions) (*Result, error)
```

Extracts content from provided HTML using optimized pooling.

**Parameters:**
- `html` (string): HTML content to parse
- `targetURL` (string): Base URL for relative links
- `opts` (*ParserOptions): Optional configuration

**Returns:**
- `*Result`: Extracted content and metadata
- `error`: Parsing error

**Example:**
```go
html := `<html><body><h1>Title</h1><p>Content...</p></body></html>`
result, err := parser.ParseHTML(html, "https://example.com", nil)
if err != nil {
    log.Fatal(err)
}
```

## Mercury Parser

The Mercury parser includes advanced features for high-performance scenarios.

### Performance Methods

#### ReturnResult

```go
func (m *Mercury) ReturnResult(result *Result)
```

Returns a result to the object pool for memory reuse.

**Parameters:**
- `result` (*Result): Result object to return to pool

**Usage:**
```go
result, err := parser.Parse("https://example.com", nil)
if err != nil {
    log.Fatal(err)
}

// Use the result...
fmt.Println(result.Title)

// Return to pool when done
parser.ReturnResult(result)
```

#### GetStats

```go
func (m *Mercury) GetStats() *PoolStats
```

Returns performance statistics for the parser instance.

**Returns:**
- `*PoolStats`: Performance metrics

`PoolStats` fields:
- TotalRequests (int64)
- PoolHits (int64)
- PoolMisses (int64)
- AverageProcessingTime (float64, ms)
- LastReset (time.Time)

**Example:**
```go
stats := parser.GetStats()
fmt.Printf("Total requests: %d\n", stats.TotalRequests)
fmt.Printf("Avg processing time: %.2fms\n", stats.AverageProcessingTime)
```

#### ResetStats

```go
func (m *Mercury) ResetStats()
```

Resets performance statistics.

## High Throughput Parser

For high-concurrency scenarios, use the `HighThroughputParser`:

```go
type HighThroughputParser struct {
    // internal fields for pooling and core parsing
}
```

### Methods

#### NewHighThroughputParser

```go
func NewHighThroughputParser(opts *ParserOptions) *HighThroughputParser
```

Creates an optimized parser for high-throughput scenarios.

#### BatchParse

```go
func (htp *HighThroughputParser) ParseBatch(urls []string, opts *ParserOptions) ([]*Result, []error)
```

Parses multiple URLs in parallel.

**Example:**
```go
htParser := parser.NewHighThroughputParser(nil)
urls := []string{
    "https://example.com/article1",
    "https://example.com/article2",
    "https://example.com/article3",
}

results, errs := htParser.ParseBatch(urls, nil)
_ = errs
for i, result := range results {
    if result != nil {
        fmt.Printf("Article %d: %s\n", i+1, result.Title)
    }
}
```

## Batch Operations

### BatchAPI

Hermes provides a BatchAPI for high-throughput processing.

```go
type BatchRequest struct {
    ID      string
    URL     string
    HTML    string
    Options *ParserOptions
    Context context.Context
    Meta    map[string]interface{}
}

type BatchResponse struct {
    ID         string
    Result     *Result
    Error      error
    Duration   time.Duration
    WorkerID   int
    ProcessedAt time.Time
}

type BatchAPIConfig struct {
    MaxWorkers        int
    QueueSize         int
    ProcessingTimeout time.Duration
    UseObjectPooling  bool
    ParserOptions     *ParserOptions
    EnableMetrics     bool
    RetryCount        int
    RetryDelay        time.Duration
}

func NewBatchAPI(config *BatchAPIConfig) *BatchAPI
func (api *BatchAPI) Start() error
func (api *BatchAPI) Stop() error
func (api *BatchAPI) Submit(req *BatchRequest) error
func (api *BatchAPI) SubmitBatch(reqs []*BatchRequest) []error
func (api *BatchAPI) GetResponse() *BatchResponse
func (api *BatchAPI) ProcessBatch(reqs []*BatchRequest) ([]*BatchResponse, error)
```

**Example:**
```go
api := parser.NewBatchAPI(nil)
defer api.Stop()
_ = api.Start()

urls := []string{"https://example.com/1", "https://example.com/2"}
reqs := make([]*parser.BatchRequest, len(urls))
for i, u := range urls {
    reqs[i] = &parser.BatchRequest{URL: u}
}
responses, err := api.ProcessBatch(reqs)
if err != nil { log.Fatal(err) }
for _, r := range responses {
    if r.Error == nil { fmt.Println(r.Result.Title) }
}
```

## Error Handling

### Error Types

The parser returns standard Go errors with additional context:

```go
// URL validation error
if err := security.ValidateURL(url); err != nil {
    return nil, fmt.Errorf("invalid URL: %w", err)
}

// Network/HTTP errors
if err := r.Create(targetURL, "", parsedURL, opts.Headers); err != nil {
    return nil, fmt.Errorf("failed to fetch content: %w", err)
}

// Extraction errors
if result.IsError() {
    return nil, fmt.Errorf("extraction failed: %s", result.Message)
}
```

### Result Error Checking

```go
func (r *Result) IsError() bool {
    return r.Error
}
```

Check if a result contains an error state.

**Example:**
```go
result, err := parser.Parse("https://example.com", nil)
if err != nil {
    log.Fatal("Network error:", err)
}

if result.IsError() {
    log.Fatal("Extraction error:", result.Message)
}

// Safe to use result
fmt.Println(result.Title)
```

## Usage Patterns

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/BumpyClock/hermes/pkg/parser"
)

func main() {
    p := parser.New()
    
    result, err := p.Parse("https://example.com/article", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    if result.IsError() {
        log.Fatal(result.Message)
    }
    
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Content: %s\n", result.Content)
}
```

### Advanced Configuration

```go
opts := &parser.ParserOptions{
    ContentType: "markdown",
    FetchAllPages: true,
    Headers: map[string]string{
        "User-Agent": "MyBot/1.0",
        "Accept": "text/html,application/xhtml+xml",
    },
}

result, err := parser.Parse("https://example.com", opts)
```

### High-Performance Batch Processing

```go
htParser := parser.NewHighThroughputParser(&parser.ParserOptions{
    ContentType: "markdown",
})

urls := []string{
    "https://example.com/1",
    "https://example.com/2",
    "https://example.com/3",
}

results, err := htParser.BatchParse(urls, nil)
for _, result := range results {
    if !result.IsError() {
        fmt.Printf("Title: %s\n", result.Title)
    }
    htParser.ReturnResult(result) // Return to pool
}
```
