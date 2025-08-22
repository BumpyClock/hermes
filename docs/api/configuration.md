# Configuration API Reference

This document covers configuration options available in Hermes, including parser options, extractor settings, and advanced configuration patterns.

Note: Environment-based configuration helpers shown in examples are illustrative. The core library exposes programmatic options via `ParserOptions`; it does not currently include built-in functions to load configuration from environment variables.

## Table of Contents

- [Parser Configuration](#parser-configuration)
- [Extractor Configuration](#extractor-configuration)
- [Content Type Configuration](#content-type-configuration)
- [HTTP Configuration](#http-configuration)
- [Performance Configuration](#performance-configuration)
- [Security Configuration](#security-configuration)
- [Environment Variables](#environment-variables)

## Parser Configuration

### ParserOptions

The main configuration structure for parser behavior.

```go
type ParserOptions struct {
    FetchAllPages   bool              // Fetch and merge multi-page articles
    Fallback        bool              // Use generic extractor as fallback
    ContentType     string            // Output format: "html", "markdown", "text"
    Headers         map[string]string // Custom HTTP headers
    CustomExtractor *CustomExtractor  // Custom extraction rules
    Extend          map[string]ExtractorFunc // Extended fields
}
```

#### Default Configuration

```go
func DefaultParserOptions() *ParserOptions {
    return &ParserOptions{
        FetchAllPages: true,
        Fallback:      true,
        ContentType:   "html",
        Headers:       nil,
        CustomExtractor: nil,
        Extend:        nil,
    }
}
```

#### Configuration Examples

**Basic Configuration:**

```go
opts := &parser.ParserOptions{
    ContentType: "markdown",
    FetchAllPages: false,
}
```

**Advanced Configuration:**

```go
opts := &parser.ParserOptions{
    FetchAllPages: true,
    Fallback: true,
    ContentType: "html",
    Headers: map[string]string{
        "User-Agent": "Hermes/1.0 (+https://example.com/bot)",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.5",
        "Accept-Encoding": "gzip, deflate",
        "DNT": "1",
        "Connection": "keep-alive",
        "Pragma": "no-cache",
        "Cache-Control": "no-cache",
    },
    Extend: map[string]parser.ExtractorFunc{
        "reading_time": func(doc *goquery.Document, url string) (interface{}, error) {
            text := doc.Find("article").Text()
            words := len(strings.Fields(text))
            readingTime := words / 200 // Assume 200 WPM
            return fmt.Sprintf("%d min read", readingTime), nil
        },
    },
}
```

### Field Options

#### FetchAllPages

Controls automatic pagination handling.

```go
// Enable multi-page detection (default)
FetchAllPages: true

// Disable multi-page detection for faster single-page extraction
FetchAllPages: false
```

**When enabled:**

- Automatically detects "next page" links
- Populates NextPageURL field if pagination detected
- **Note: Automatic fetching and merging of subsequent pages is not yet implemented**

**When disabled:**

- Only processes the initial page
- Faster extraction for single-page articles
- NextPageURL field will be empty even if pagination exists

**Current Limitations:**

- Multi-page content fetching and merging is not functional
- Only next page URL detection is implemented
- See TODO items in codebase for planned implementation

#### Fallback

Controls fallback to generic extractor when custom extractors fail.

```go
// Enable fallback (default) - recommended for production
Fallback: true

// Disable fallback - custom extractors only
Fallback: false
```

**When enabled:**

- Uses custom extractor if available
- Falls back to generic algorithm if custom fails
- Ensures content extraction even for unsupported sites

**When disabled:**

- Only uses custom extractors
- Returns empty result if no custom extractor available
- Useful for testing custom extractors

#### ContentType

Specifies the output format for extracted content.

```go
// Available options
ContentType: "html"     // Clean HTML (default)
ContentType: "markdown" // Markdown format
ContentType: "text"     // Plain text
ContentType: "json"     // JSON structure
```

**Format Examples:**

**HTML Output:**

```html
<h2>Section Title</h2>
<p>Article content with <strong>emphasis</strong> and <a href="...">links</a>.</p>
<ul>
  <li>List item 1</li>
  <li>List item 2</li>
</ul>
```

**Markdown Output:**

```markdown
## Section Title

Article content with **emphasis** and [links](...).

- List item 1
- List item 2
```

**Text Output:**

```
Section Title

Article content with emphasis and links.

- List item 1
- List item 2
```

## Extractor Configuration

### ExtractorOptions

Configuration for individual extractor operations.

```go
type ExtractorOptions struct {
    URL         string
    HTML        string
    MetaCache   map[string]string
    Fallback    bool
    ContentType string
}
```

#### Default Extractor Options

```go
func DefaultExtractorOptions() *ExtractorOptions {
    return &ExtractorOptions{
        Fallback:    true,
        ContentType: "html",
        MetaCache:   make(map[string]string),
    }
}
```

### Generic Extractor Configuration

```go
type ExtractorOptions struct {
    StripUnlikelyCandidates bool // Remove unlikely content elements
    WeightNodes             bool // Apply content scoring algorithm  
    CleanConditionally      bool // Apply conditional cleaning rules
}
```

#### Extraction Strategy Configuration

**Strict Extraction (High Quality):**

```go
opts := ExtractorOptions{
    StripUnlikelyCandidates: true,
    WeightNodes:             true,
    CleanConditionally:      true,
}
```

**Permissive Extraction (More Content):**

```go
opts := ExtractorOptions{
    StripUnlikelyCandidates: false,
    WeightNodes:             false,
    CleanConditionally:      false,
}
```

**Balanced Extraction (Recommended):**

```go
opts := ExtractorOptions{
    StripUnlikelyCandidates: true,
    WeightNodes:             true,
    CleanConditionally:      false,
}
```

## Content Type Configuration

### HTML Configuration

```go
ContentType: "html"
```

**Features:**

- Preserves HTML structure and formatting
- Includes links, images, lists, and emphasis
- Applies content cleaning and normalization
- Best for web display or further HTML processing

**Cleaning Options:**

```go
// Control HTML cleaning behavior
defaultCleaner := true  // Apply standard HTML cleaning
cleanConditionally := true  // Apply conditional cleaning rules
```

### Markdown Configuration

```go
ContentType: "markdown"
```

**Features:**

- Converts HTML to clean Markdown
- Preserves formatting and structure
- Ideal for documentation and text processing
- Includes metadata header

**Example Output:**

```markdown
# Article Title

**Author:** John Doe  
**Date:** 2024-01-15T10:30:00Z  
**URL:** https://example.com/article

## Section Heading

Article content with **bold text** and [links](https://example.com).

- List item one
- List item two

> Blockquote content
```

### Text Configuration

```go
ContentType: "text"
```

**Features:**

- Plain text output with minimal formatting
- Removes all HTML tags and structure
- Preserves line breaks and basic structure
- Smallest output size

### JSON Configuration

```go
ContentType: "json"
```

**Features:**

- Complete structured data export
- Includes all extracted fields and metadata
- Programmatic access to all content
- Ideal for API responses and data processing

## HTTP Configuration

### Custom Headers

Configure HTTP headers for web requests.

```go
Headers: map[string]string{
    "User-Agent": "Hermes/1.0 (+https://example.com/bot)",
    "Accept": "text/html,application/xhtml+xml",
    "Accept-Language": "en-US,en;q=0.5",
    "Referer": "https://google.com/",
    "Cookie": "session=abc123; preferences=xyz789",
}
```

#### Common Header Patterns

**Bot Identification:**

```go
Headers: map[string]string{
    "User-Agent": "Hermes Bot/1.0 (+https://yoursite.com/bot-info)",
    "From": "bot@yoursite.com",
}
```

**Browser Simulation:**

```go
Headers: map[string]string{
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    "Accept-Language": "en-US,en;q=0.5",
    "Accept-Encoding": "gzip, deflate",
    "DNT": "1",
    "Connection": "keep-alive",
    "Upgrade-Insecure-Requests": "1",
}
```

**API Access:**

```go
Headers: map[string]string{
    "Authorization": "Bearer " + token,
    "X-API-Key": apiKey,
    "Content-Type": "application/json",
}
```

### Request Configuration

```go
type RequestConfig struct {
    Timeout       time.Duration // Request timeout
    MaxRedirects  int          // Maximum redirects to follow
    RetryAttempts int          // Number of retry attempts
    RetryDelay    time.Duration // Delay between retries
}
```

**Default Configuration:**

```go
RequestConfig{
    Timeout:       30 * time.Second,
    MaxRedirects:  10,
    RetryAttempts: 3,
    RetryDelay:    1 * time.Second,
}
```

## Performance Configuration

### High Throughput Configuration

```go
type HighThroughputConfig struct {
    MaxConcurrency    int           // Maximum concurrent operations
    PoolSize         int           // Object pool size
    WorkerCount      int           // Number of worker goroutines
    QueueSize        int           // Worker queue size
    EnableProfiling  bool          // Enable performance profiling
}
```

**Production Configuration:**

```go
config := HighThroughputConfig{
    MaxConcurrency:   100,
    PoolSize:        1000,
    WorkerCount:     runtime.NumCPU() * 2,
    QueueSize:       10000,
    EnableProfiling: false,
}
```

**Development Configuration:**

```go
config := HighThroughputConfig{
    MaxConcurrency:   10,
    PoolSize:        100,
    WorkerCount:     4,
    QueueSize:       1000,
    EnableProfiling: true,
}
```

### Memory Management

```go
type MemoryConfig struct {
    EnablePooling     bool // Enable object pooling
    PoolMaxSize      int  // Maximum pool size
    GCPercent        int  // Garbage collection percentage
    MaxMemoryUsage   int64 // Maximum memory usage in bytes
}
```

### Batch Processing Configuration

```go
type BatchOptions struct {
    Concurrency     int           // Number of concurrent requests
    Timeout         time.Duration // Timeout per request
    RetryAttempts   int          // Retry attempts for failed requests
    ProgressCallback func(completed, total int) // Progress callback
}
```

**Example:**

```go
batchOpts := &parser.BatchOptions{
    Concurrency: 10,
    Timeout: 30 * time.Second,
    RetryAttempts: 2,
    ProgressCallback: func(completed, total int) {
        fmt.Printf("Progress: %d/%d (%.1f%%)\n", 
            completed, total, float64(completed)/float64(total)*100)
    },
}
```

## Security Configuration

### URL Validation

```go
type SecurityConfig struct {
    AllowedProtocols []string // Allowed URL protocols
    AllowedDomains   []string // Allowed domains (whitelist)
    BlockedDomains   []string // Blocked domains (blacklist)
    MaxURLLength     int      // Maximum URL length
    ValidateSSL      bool     // Validate SSL certificates
}
```

**Secure Configuration:**

```go
config := SecurityConfig{
    AllowedProtocols: []string{"https"},
    AllowedDomains:   []string{"example.com", "*.example.com"},
    BlockedDomains:   []string{"malicious.com", "spam.com"},
    MaxURLLength:     2048,
    ValidateSSL:      true,
}
```

### Content Sanitization

```go
type SanitizationConfig struct {
    RemoveScripts    bool     // Remove script tags
    RemoveStyles     bool     // Remove style tags
    AllowedTags      []string // Allowed HTML tags
    AllowedAttrs     []string // Allowed HTML attributes
    MaxContentLength int      // Maximum content length
}
```

## Environment Variables

Hermes supports configuration via environment variables:

### Core Settings

```bash
# Parser settings
HERMES_CONTENT_TYPE=markdown
HERMES_FETCH_ALL_PAGES=true
HERMES_FALLBACK=true

# HTTP settings
HERMES_TIMEOUT=30s
HERMES_MAX_REDIRECTS=10
HERMES_USER_AGENT="Hermes/1.0"

# Performance settings
HERMES_MAX_CONCURRENCY=100
HERMES_POOL_SIZE=1000
HERMES_WORKER_COUNT=8

# Security settings
HERMES_VALIDATE_SSL=true
HERMES_MAX_URL_LENGTH=2048
```

### Loading Environment Configuration

```go
func LoadConfigFromEnv() *ParserOptions {
    opts := DefaultParserOptions()
    
    if contentType := os.Getenv("HERMES_CONTENT_TYPE"); contentType != "" {
        opts.ContentType = contentType
    }
    
    if fetchAll := os.Getenv("HERMES_FETCH_ALL_PAGES"); fetchAll == "false" {
        opts.FetchAllPages = false
    }
    
    if fallback := os.Getenv("HERMES_FALLBACK"); fallback == "false" {
        opts.Fallback = false
    }
    
    if userAgent := os.Getenv("HERMES_USER_AGENT"); userAgent != "" {
        if opts.Headers == nil {
            opts.Headers = make(map[string]string)
        }
        opts.Headers["User-Agent"] = userAgent
    }
    
    return opts
}
```

## Configuration Validation

### Validation Functions

```go
func ValidateParserOptions(opts *ParserOptions) error {
    if opts == nil {
        return errors.New("parser options cannot be nil")
    }
    
    // Validate content type
    validTypes := []string{"html", "markdown", "text", "json"}
    if !contains(validTypes, opts.ContentType) {
        return fmt.Errorf("invalid content type: %s", opts.ContentType)
    }
    
    // Validate headers
    if opts.Headers != nil {
        for key, value := range opts.Headers {
            if key == "" || value == "" {
                return errors.New("header key and value cannot be empty")
            }
        }
    }
    
    return nil
}
```

### Configuration Examples

**Complete Production Configuration:**

```go
config := &parser.ParserOptions{
    FetchAllPages: true,
    Fallback: true,
    ContentType: "html",
    Headers: map[string]string{
        "User-Agent": "Hermes/1.0 (+https://yoursite.com/bot)",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.5",
        "Accept-Encoding": "gzip, deflate",
        "DNT": "1",
        "Connection": "keep-alive",
    },
    Extend: map[string]parser.ExtractorFunc{
        "social_shares": extractSocialShares,
        "reading_time": calculateReadingTime,
        "word_count": countWords,
    },
}

// Validate configuration
if err := ValidateParserOptions(config); err != nil {
    log.Fatal("Invalid configuration:", err)
}

// Create parser with configuration
parser := parser.New(config)
```

**Minimal Configuration:**

```go
config := &parser.ParserOptions{
    ContentType: "markdown",
}

parser := parser.New(config)
```

**Testing Configuration:**

```go
config := &parser.ParserOptions{
    FetchAllPages: false,
    Fallback: false,
    ContentType: "html",
    Headers: map[string]string{
        "User-Agent": "Hermes-Test/1.0",
    },
}
```
