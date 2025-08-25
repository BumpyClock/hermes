# Architecture Overview

This document provides a comprehensive overview of Hermes' architecture, design principles, and system components.

## Table of Contents

- [Design Principles](#design-principles)
- [System Architecture](#system-architecture)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Performance Architecture](#performance-architecture)
- [Security Architecture](#security-architecture)
- [Extensibility](#extensibility)

## Design Principles

Hermes is built on several key design principles that guide architectural decisions:

### 1. Performance First

- **Zero-allocation processing** where possible using object pooling
- **Concurrent processing** with configurable worker pools
- **Streaming extraction** for large documents
- **Memory-efficient DOM operations** with selective parsing

### 2. Compatibility

- **100% API compatibility** with JavaScript Postlight Parser
- **Identical extraction results** for existing custom extractors
- **Same CLI interface** and command structure
- **Drop-in replacement** for existing workflows

### 3. Reliability

- **Graceful error handling** with cascading fallbacks
- **Robust parsing** that handles malformed HTML
- **Comprehensive testing** with 150+ site fixtures
- **Production-ready** with extensive validation

### 4. Extensibility

- **Plugin architecture** for custom extractors
- **Configurable processing** pipeline
- **Custom field extraction** capabilities
- **Transform functions** for content modification

### 5. Simplicity

- **Clear separation** of concerns between components
- **Minimal dependencies** with carefully vetted libraries
- **Simple APIs** that hide complexity
- **Self-documenting code** with comprehensive comments

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        User Interface                       │
├─────────────────────────────────────────────────────────────┤
│  CLI Tool          │  Go Library API   │  HTTP Service     │
│  (cmd/parser)      │  (pkg/parser)     │  (optional)       │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                     Parser Layer                           │
├─────────────────────────────────────────────────────────────┤
│  Hermes Parser    │  High Throughput  │  Batch API        │
│  (Single requests) │  Parser (Pooled)  │  (Bulk ops)       │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                   Extraction Layer                         │
├─────────────────────────────────────────────────────────────┤
│  Custom Extractors │  Generic Extractors │  Field Extractors│
│  (Site-specific)   │  (Algorithm-based)  │  (Individual)    │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                    Processing Layer                        │
├─────────────────────────────────────────────────────────────┤
│  Content Cleaners  │  Text Processors   │  Format Converters│
│  (Sanitization)    │  (Normalization)   │  (Output formats) │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                    Resource Layer                          │
├─────────────────────────────────────────────────────────────┤
│  HTTP Client       │  DOM Parser        │  Encoding Handler │
│  (Fetching)        │  (goquery)         │  (chardet)        │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                     │
├─────────────────────────────────────────────────────────────┤
│  Object Pools      │  Worker Pools      │  Cache Layer      │
│  (Memory reuse)    │  (Concurrency)     │  (Performance)    │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Parser Layer

#### Hermes Parser

The main parser implementation providing the primary interface:

```go
type Hermes struct {
    options  ParserOptions
    htParser *HighThroughputParser
}
```

**Responsibilities:**
- API entry point for single requests
- Configuration management
- Error handling and validation
- Integration with high-throughput parser

**Key Methods:**
- `Parse(url, options)` - Extract content from URL
- `ParseHTML(html, url, options)` - Extract from HTML string
- `ReturnResult(result)` - Return result to object pool

#### High Throughput Parser

Optimized parser for high-concurrency scenarios:

```go
type HighThroughputParser struct {
    options    *ParserOptions
    resultPool *ObjectPool[Result]
    workerPool *WorkerPool
}
```

**Responsibilities:**
- Object pooling for memory efficiency
- Worker pool management
- Batch processing coordination
- Performance monitoring

**Key Features:**
- Memory reuse through object pooling
- Configurable concurrency limits
- Performance statistics tracking
- Resource cleanup management

### 2. Extraction Layer

#### Custom Extractors

Site-specific extractors optimized for particular domains:

```go
type CustomExtractor struct {
    Domain        string
    Title         *FieldExtractor
    Author        *FieldExtractor
    Content       *ContentExtractor
    DatePublished *FieldExtractor
    // ... other fields
}
```

**Responsibilities:**
- Site-specific content extraction
- CSS selector-based field extraction
- Content transformation and cleaning
- Multi-domain support

**Registry Management:**
```go
type ExtractorRegistry struct {
    extractors map[string]*CustomExtractor
}
```

#### Generic Extractors

Algorithm-based extractors for fallback scenarios:

```go
type GenericContentExtractor struct {
    DefaultOpts ExtractorOptions
}
```

**Responsibilities:**
- Content scoring algorithm
- Cascading extraction strategies
- DOM analysis and candidate selection
- Fallback content extraction

**Extraction Strategy:**
1. **Strict extraction** with full algorithm
2. **Progressive relaxation** of constraints
3. **Minimal extraction** as last resort

#### Field Extractors

Specialized extractors for individual content fields:

```go
type FieldExtractor struct {
    Selectors      []interface{}
    AllowMultiple  bool
    DefaultCleaner bool
}
```

**Field Types:**
- **Title Extraction** - Headlines and titles
- **Author Extraction** - Bylines and author information
- **Date Extraction** - Publication dates with parsing
- **Image Extraction** - Lead images and media
- **Content Extraction** - Main article content

### 3. Processing Layer

#### Content Cleaners

Specialized cleaners for different content types:

```go
// pkg/cleaners/
├── title.go          // Title cleaning and normalization
├── author.go         // Author name processing
├── content.go        // Main content cleaning
├── date_published.go // Date parsing and validation
└── lead_image_url.go // Image URL processing
```

**Cleaning Pipeline:**
1. **HTML sanitization** - Remove scripts, dangerous content
2. **Content normalization** - Fix encoding, spacing
3. **Structure cleaning** - Remove navigation, ads
4. **Format conversion** - Convert to target format

#### Text Processors

Text manipulation utilities:

```go
// pkg/utils/text/
├── normalize_spaces.go  // Whitespace normalization
├── extract_from_url.go  // URL-based extraction
├── date.go             // Date parsing utilities
├── excerpt_content.go  // Summary generation
└── encoding.go         // Character encoding
```

### 4. Resource Layer

#### HTTP Client

Robust HTTP client with retries and timeout handling:

```go
type HTTPClient struct {
    client      *http.Client
    maxRetries  int
    timeout     time.Duration
    userAgent   string
}
```

**Features:**
- Automatic retry with exponential backoff
- Configurable timeouts and redirects
- Custom headers and authentication
- Response caching capabilities

#### DOM Parser

HTML parsing and manipulation using goquery:

```go
// pkg/resource/
├── resource.go    // Main resource handling
├── dom.go         // DOM manipulation utilities
├── encoding.go    // Character encoding detection
└── http.go        // HTTP client implementation
```

**Capabilities:**
- jQuery-like DOM manipulation
- CSS selector-based extraction
- HTML parsing and validation
- Character encoding detection

### 5. Infrastructure Layer

#### Object Pools

Memory-efficient object reuse:

```go
type ObjectPool[T any] struct {
    pool    sync.Pool
    factory func() *T
    stats   *PoolStats
}
```

**Benefits:**
- Reduced garbage collection pressure
- Consistent memory usage patterns
- Performance monitoring
- Configurable pool sizes

#### Worker Pools

Concurrent processing management:

```go
type WorkerPool struct {
    workers     int
    jobQueue    chan Job
    resultQueue chan Result
    wg          sync.WaitGroup
}
```

**Features:**
- Configurable worker count
- Job queue management
- Graceful shutdown
- Backpressure handling

## Data Flow

### 1. Request Processing Flow

```
URL Request
    │
    ▼
┌─────────────┐
│ URL         │
│ Validation  │
└─────────────┘
    │
    ▼
┌─────────────┐
│ HTTP        │
│ Fetch       │
└─────────────┘
    │
    ▼
┌─────────────┐
│ HTML        │
│ Parsing     │
└─────────────┘
    │
    ▼
┌─────────────┐
│ Extractor   │
│ Selection   │
└─────────────┘
    │
    ▼
┌─────────────┐
│ Content     │
│ Extraction  │
└─────────────┘
    │
    ▼
┌─────────────┐
│ Content     │
│ Cleaning    │
└─────────────┘
    │
    ▼
┌─────────────┐
│ Format      │
│ Conversion  │
└─────────────┘
    │
    ▼
Result Object
```

### 2. Extractor Selection Logic

```
Parse Request
    │
    ▼
┌─────────────────┐
│ Extract Domain  │
│ from URL        │
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ Check Custom    │
│ Extractor       │
│ Registry        │
└─────────────────┘
    │
    ├── Found ────────────┐
    │                     ▼
    │              ┌─────────────────┐
    │              │ Use Custom      │
    │              │ Extractor       │
    │              └─────────────────┘
    │                     │
    └── Not Found ────────┼─────────┐
                          │         ▼
                          │  ┌─────────────────┐
                          │  │ Use Generic     │
                          │  │ Extractor       │
                          │  └─────────────────┘
                          │         │
                          ▼         ▼
                    ┌─────────────────┐
                    │ Extraction      │
                    │ Results         │
                    └─────────────────┘
```

### 3. Content Processing Pipeline

```
Raw HTML
    │
    ▼
┌─────────────────┐
│ DOM Parsing     │
│ (goquery)       │
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ Field           │
│ Extraction      │
└─────────────────┘
    │
    ├── Title ──────┐
    ├── Author ─────┤
    ├── Content ────┤
    ├── Date ───────┤
    └── Images ─────┘
                    │
                    ▼
            ┌─────────────────┐
            │ Field Cleaning  │
            │ & Validation    │
            └─────────────────┘
                    │
                    ▼
            ┌─────────────────┐
            │ Content         │
            │ Transformation  │
            └─────────────────┘
                    │
                    ▼
            ┌─────────────────┐
            │ Format          │
            │ Conversion      │
            └─────────────────┘
                    │
                    ▼
              Final Result
```

## Performance Architecture

### 1. Memory Management

#### Object Pooling Strategy

```go
// Result object pooling
var resultPool = &ObjectPool[Result]{
    pool: sync.Pool{
        New: func() interface{} {
            return &Result{}
        },
    },
}

// Document pooling for reuse
var documentPool = &ObjectPool[goquery.Document]{
    pool: sync.Pool{
        New: func() interface{} {
            return &goquery.Document{}
        },
    },
}
```

**Benefits:**
- 50% reduction in memory allocations
- Consistent memory usage patterns
- Reduced garbage collection pressure
- Improved cache locality

#### Memory Usage Patterns

```
Memory Usage Over Time (Without Pooling)
┌─────────────────────────────────────────┐
│    ▲                                    │
│   ▲ ▲                                   │
│  ▲   ▲     ▲                           │
│ ▲     ▲   ▲ ▲                          │
│▲       ▲ ▲   ▲    ▲                    │
│         ▲     ▲  ▲ ▲                   │
│                ▲▲   ▲                  │
└─────────────────────────────────────────┘
  Sawtooth pattern with GC spikes

Memory Usage Over Time (With Pooling)
┌─────────────────────────────────────────┐
│████████████████████████████████████████ │
│████████████████████████████████████████ │
│████████████████████████████████████████ │
│████████████████████████████████████████ │
└─────────────────────────────────────────┘
  Stable, predictable memory usage
```

### 2. Concurrency Architecture

#### Worker Pool Design

```go
type WorkerPool struct {
    workers     int
    jobs        chan Job
    results     chan Result
    quit        chan bool
    wg          sync.WaitGroup
}
```

**Scaling Strategy:**
- **CPU-bound tasks**: Workers = CPU cores
- **I/O-bound tasks**: Workers = 2-4x CPU cores
- **Network requests**: Configurable based on target capacity

#### Concurrency Patterns

```
Single Request Processing:
┌──────┐    ┌─────────┐    ┌────────┐
│Client│───▶│ Parser  │───▶│ Result │
└──────┘    └─────────┘    └────────┘

Batch Request Processing:
┌──────┐    ┌─────────┐    ┌─────────┐    ┌────────┐
│Client│───▶│Scheduler│───▶│Worker[1]│───▶│Results │
└──────┘    └─────────┘    ├─────────┤    └────────┘
                           │Worker[2]│
                           ├─────────┤
                           │Worker[N]│
                           └─────────┘
```

### 3. Caching Strategy

#### Multi-Level Caching

```go
type CacheLayer struct {
    metaCache     map[string]string    // Meta tag cache
    domCache      map[string]*Document // Parsed DOM cache
    resultCache   map[string]*Result   // Final result cache
    extractorCache map[string]*CustomExtractor // Extractor cache
}
```

**Cache Hierarchy:**
1. **Meta tag cache** - Extracted meta information
2. **DOM cache** - Parsed document structures
3. **Result cache** - Complete extraction results
4. **Extractor cache** - Compiled custom extractors

Note: The above illustrates a conceptual multi-layer cache. In the current codebase, Hermes implements an extractor loader cache and DOM operation caches (see `pkg/extractors/loader.go` and `pkg/cache/helpers.go`). A full meta/result cache is planned but not yet implemented.

## Security Architecture

### 1. Input Validation

#### URL Validation Pipeline

```go
func ValidateURL(rawURL string) error {
    // 1. Parse URL structure
    parsedURL, err := url.Parse(rawURL)
    if err != nil {
        return fmt.Errorf("invalid URL format: %w", err)
    }
    
    // 2. Check protocol whitelist
    if !isAllowedProtocol(parsedURL.Scheme) {
        return fmt.Errorf("protocol not allowed: %s", parsedURL.Scheme)
    }
    
    // 3. Check domain blacklist
    if isBlockedDomain(parsedURL.Host) {
        return fmt.Errorf("domain blocked: %s", parsedURL.Host)
    }
    
    // 4. Validate URL length
    if len(rawURL) > MaxURLLength {
        return fmt.Errorf("URL too long: %d > %d", len(rawURL), MaxURLLength)
    }
    
    return nil
}
```

#### Content Sanitization

```go
// HTML sanitization pipeline
func SanitizeHTML(html string) string {
    // 1. Parse HTML safely
    doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
    
    // 2. Remove dangerous elements
    doc.Find("script, style, iframe, object, embed").Remove()
    
    // 3. Sanitize attributes
    sanitizeAttributes(doc)
    
    // 4. Apply content security policies
    applyCsp(doc)
    
    return renderHTML(doc)
}
```

### 2. Resource Limits

#### Request Limits

```go
type ResourceLimits struct {
    MaxContentLength  int64         // Maximum response size
    MaxRedirects      int           // Maximum redirect follows
    RequestTimeout    time.Duration // Maximum request time
    MaxConcurrency    int           // Maximum concurrent requests
    RateLimit         rate.Limiter  // Request rate limiting
}
```

#### Memory Limits

```go
type MemoryLimits struct {
    MaxPoolSize       int    // Maximum object pool size
    MaxCacheSize      int64  // Maximum cache memory usage
    GCPercent         int    // Garbage collection percentage
    MaxDocumentSize   int64  // Maximum document size to parse
}
```

## Extensibility

### 1. Plugin Architecture (planned)

#### Custom Extractor Interface

```go
type ExtractorPlugin interface {
    GetDomains() []string
    Extract(doc *goquery.Document, url string) (*Result, error)
    Configure(config map[string]interface{}) error
}
```

#### Registration System

```go
type PluginRegistry struct {
    plugins map[string]ExtractorPlugin
    hooks   map[string][]HookFunc
}

func (r *PluginRegistry) Register(name string, plugin ExtractorPlugin) {
    r.plugins[name] = plugin
    
    // Auto-register for supported domains
    for _, domain := range plugin.GetDomains() {
        r.registerForDomain(domain, plugin)
    }
}
```

### 2. Hook System (planned)

#### Processing Hooks

```go
type HookType string

const (
    PreFetch    HookType = "pre_fetch"
    PostFetch   HookType = "post_fetch"
    PreExtract  HookType = "pre_extract"
    PostExtract HookType = "post_extract"
    PreClean    HookType = "pre_clean"
    PostClean   HookType = "post_clean"
)

type HookFunc func(ctx *Context) error

type Context struct {
    URL      string
    HTML     string
    Document *goquery.Document
    Result   *Result
    Options  *ParserOptions
}
```

#### Hook Registration

```go
func RegisterHook(hookType HookType, fn HookFunc) {
    hooks[hookType] = append(hooks[hookType], fn)
}

func ExecuteHooks(hookType HookType, ctx *Context) error {
    for _, hook := range hooks[hookType] {
        if err := hook(ctx); err != nil {
            return fmt.Errorf("hook %s failed: %w", hookType, err)
        }
    }
    return nil
}
```

### 3. Configuration System (planned)

#### Hierarchical Configuration

```go
type Config struct {
    Global    GlobalConfig              // Global settings
    Domains   map[string]DomainConfig   // Domain-specific settings
    Plugins   map[string]PluginConfig   // Plugin configurations
    Overrides map[string]interface{}    // Runtime overrides
}
```

**Configuration Sources (in priority order):**
1. Runtime options passed to Parse()
2. Environment variables
3. Configuration files
4. Default values

This architecture provides a solid foundation for high-performance, reliable, and extensible web content extraction while maintaining compatibility with existing tools and workflows.
