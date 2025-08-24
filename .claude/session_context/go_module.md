# Hermes Go Module Refactoring Session

## Session Overview

**Branch**: `aditya/go-module-refactor` (created and active)
**Goal**: Transform Hermes from a CLI tool into a best-in-class Go library while maintaining the CLI functionality
**Date**: 2025-08-24

## Problem Statement

The current Hermes parser works great as a CLI but needs improvements for use as a Go module:

- API is buried in `pkg/parser` with complex internals exposed
- Mix of optimization layers (batch processing, object pooling) complicates simple use cases
- No clean, reusable client pattern for API servers
- Batch processing logic belongs in CLI, not library
- Context not properly threaded through the codebase
- Global HTTP client prevents proper resource management
- Generic errors make programmatic handling difficult

## Core Design Decisions

### Versioning Strategy (v1)

- Keep the module path as `github.com/BumpyClock/hermes` (no `/v2`)
- Make breaking changes now (no external consumers), document them clearly
- Immediately internalize implementation (`pkg/*` → `internal/*`) and provide minimal public API in root package
- Update CLI to use new root package; remove references to `pkg/*` from application code

### Philosophy

- **Single Responsibility**: Parse web content excellently (ONLY parsing, not orchestration)
- **Unix Philosophy**: Do one thing well
- **Lean API**: Minimal surface area, maximum utility
- **Go Idiomatic**: Follow Go conventions, not Java/JS patterns
- **Composability**: Simple building block that callers can orchestrate as needed

### Key Requirements (from discussion - FINAL)

1. Reusable client that can be initialized once and used across requests
2. Parse single URLs efficiently (caller handles batching/concurrency)
3. Standard Go return pattern: `(*Result, error)`
4. No caching, retries, or deduplication (caller's responsibility)
5. Context support for cancellation/timeouts (properly threaded)
6. Typed error handling for programmatic use
7. Mockable interface for testing
8. Client-owned HTTP client for connection pooling
9. Unix philosophy: do ONE thing well

## Critical Technical Improvements

### 1. Context Plumbing End-to-End

**Problem**: Current code creates its own contexts internally, preventing proper cancellation

**Solution**: Thread context through entire call chain
- `pkg/resource/http.go`: Update `Get`, `GetWithRetry`, `doRequest` to accept `ctx`
- `pkg/resource/fetch.go`: Add `ctx` to `FetchResource`, remove `getGlobalHTTPClient`
- `pkg/resource/resource.go`: Update `Create` method to accept context
- `pkg/utils/security/dns.go`: Switch from `net.LookupIP` to `net.Resolver.LookupIPAddr(ctx, host)`

### 2. Client-Owned HTTP Client

**Problem**: Global singletons prevent proper resource management

**Solution**: Client owns its HTTP client
- Inject via `WithHTTPClient(*http.Client)` or `WithTransport(http.RoundTripper)`
- Keep connection pooling in transport
- Re-enable HTTP/2 by default (remove workaround)
- Let callers override if needed

### 3. Typed Error Model

**Problem**: Generic errors make programmatic handling difficult

**Solution**: Introduce structured errors

```go
type ParseError struct {
    Code ErrorCode  // InvalidURL, Fetch, Timeout, SSRF, Extract
    URL  string
    Op   string     // Operation that failed
    Err  error      // Underlying error
}

func (e *ParseError) Error() string
func (e *ParseError) Unwrap() error

// Sentinel error codes
const (
    ErrInvalidURL ErrorCode = iota
    ErrFetch
    ErrTimeout
    ErrSSRF
    ErrExtract
)
```

### 4. Clean Result Type

**Problem**: Internal types leak through public API

**Solution**: Define explicit public Result
- Map from internal `parser.Result` to public `hermes.Result`
- Hide internal fields
- Keep helpful methods like `FormatMarkdown()`
- Ensure no internal type exposure

### 5. SSRF Protection Options

**Problem**: Hard-coded SSRF checks may be too restrictive

**Solution**: Configurable security
- Add `WithAllowPrivateNetworks(bool)` option
- Default to secure (false)
- Let trusted environments relax checks

## Final API Design

```go
package hermes

// Client - thread-safe, reusable, single responsibility
type Client struct {
    httpClient           *http.Client
    userAgent            string
    timeout              time.Duration
    allowPrivateNetworks bool
}

// Constructor with optional configuration
func New(opts ...Option) *Client

// Parse ONE URL - standard Go pattern
func (c *Client) Parse(ctx context.Context, url string) (*Result, error)

// ParseHTML for pre-fetched content
func (c *Client) ParseHTML(ctx context.Context, html, url string) (*Result, error)

// Options for configuration (functional options pattern)
func WithHTTPClient(client *http.Client) Option
func WithTransport(rt http.RoundTripper) Option
func WithTimeout(d time.Duration) Option
func WithUserAgent(ua string) Option
func WithAllowPrivateNetworks(allow bool) Option

// Clean public Result type
type Result struct {
    URL           string     `json:"url"`
    Title         string     `json:"title"`
    Content       string     `json:"content"`
    Author        string     `json:"author,omitempty"`
    DatePublished *time.Time `json:"date_published,omitempty"`
    LeadImageURL  string     `json:"lead_image_url,omitempty"`
    Dek           string     `json:"dek,omitempty"`
    Domain        string     `json:"domain"`
    Excerpt       string     `json:"excerpt,omitempty"`
    WordCount     int        `json:"word_count"`
    Direction     string     `json:"direction,omitempty"`
    TotalPages    int        `json:"total_pages,omitempty"`
    RenderedPages int        `json:"rendered_pages,omitempty"`
    SiteName      string     `json:"site_name,omitempty"`
}

func (r *Result) FormatMarkdown() string

// Typed error for programmatic handling
type ParseError struct {
    Code ErrorCode
    URL  string
    Op   string
    Err  error
}

// Parser interface for mocking
type Parser interface {
    Parse(ctx context.Context, url string) (*Result, error)
    ParseHTML(ctx context.Context, html, url string) (*Result, error)
}
```

## Phased Implementation Plan

### Phase A: Add New Public API (Safe Addition)

**Status**: In Progress

**Goal**: Create root package without breaking anything

**Files to create:**
- `client.go` - Client struct with HTTP client ownership
- `result.go` - Public Result type (maps from internal)
- `errors.go` - ParseError and error codes
- `options.go` - Functional options
- `parser.go` - Parser interface for mocking

**Key implementation details:**
- Wrap existing `pkg/parser` functionality
- Map internal types to public types
- No behavior changes yet

### Phase B: Context Plumbing (Internal Fix)

**Status**: Pending

**Goal**: Fix context handling throughout codebase

**Files to modify:**
- `pkg/resource/http.go` - Thread context through HTTP calls
- `pkg/resource/fetch.go` - Accept context parameter
- `pkg/resource/resource.go` - Pass context to Create
- `pkg/utils/security/dns.go` - Use context-aware DNS resolution
- Update all callers to pass context

**Verification**: Run existing tests to ensure no regressions

### Phase C: Internalize Implementation (Breaking Change)

**Status**: Pending

**Goal**: Hide implementation details

**Actions:**
1. Move `pkg/*` → `internal/*`
2. Update all import paths
3. Remove public access to internal packages
4. Update CLI to use root package API

**Note**: This is the breaking change, but safe since no external consumers

### Phase D: Remove Orchestration Code (Cleanup)

**Status**: Pending

**Goal**: Simplify by removing unnecessary complexity

**Safe removal after verification:**
- `internal/parser/batch_api.go`
- `internal/parser/worker_pool.go`
- `internal/parser/object_pool.go`
- `internal/parser/streaming.go` (verify no regressions first)

**Note**: Run memory/performance tests before removing streaming

### Phase E: Update CLI (Refactor)

**Status**: Pending

**Goal**: CLI uses new public API

**Changes to `cmd/parser/main.go`:**
- Use `hermes.New()` with options
- Implement own concurrency with semaphore pattern
- Keep timing, progress reporting, output formatting
- Example concurrent implementation for batch processing

**CLI retains:**
- Batch processing orchestration (using semaphore)
- Progress reporting
- Timing measurements
- Output file handling
- Multiple format outputs

### Phase F: Documentation & Examples

**Status**: Pending

**Goal**: Make library approachable

**Files to create:**
- `doc.go` - Package documentation
- `example_test.go` - Testable examples
- `examples/basic/main.go` - Simple single-URL parsing
- `examples/concurrent/main.go` - Semaphore pattern for batching
- `examples/custom-client/main.go` - Custom HTTP client injection
- `examples/api-server/main.go` - HTTP handler integration
- Updated `README.md` with library section

### Phase G: Comprehensive Testing

**Status**: Pending

**Goal**: Ensure quality and prevent regressions

**Test coverage needed:**
- Context cancellation behavior
- HTTP client injection
- Timeout handling
- Error type assertions
- Memory usage (especially after removing streaming)
- Fixture-based extraction tests (keep existing)
- Mock client implementation

## Decisions Made

### What We're Including

- ✅ Context support (properly threaded throughout)
- ✅ Client-owned HTTP client (no globals)
- ✅ Connection pooling (HTTP client reuse for efficiency)
- ✅ Typed error handling (ParseError with codes)
- ✅ Mockable interface (testing)
- ✅ Thread-safe client (reusable across goroutines)
- ✅ Functional options (2024 Go best practice)
- ✅ SSRF protection options (configurable)

### What We're Excluding

- ❌ Internal concurrency (caller's responsibility)
- ❌ Batch processing in library (move to CLI)
- ❌ Structured batch responses (single URL = single response)
- ❌ Streaming API (evaluate before removal)
- ❌ Middleware/plugins (unnecessary complexity)
- ❌ Built-in caching (caller's responsibility)
- ❌ Built-in retries (caller's responsibility)
- ❌ Complex builders (use functional options)
- ❌ Runtime custom extractors (defer to later)
- ❌ Format negotiation (use options)

## Example Usage (Target State)

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/BumpyClock/hermes"
)

func main() {
    // Initialize once
    client := hermes.New(
        hermes.WithTimeout(30*time.Second),
        hermes.WithUserAgent("MyApp/1.0"),
    )
    
    // Parse single URL - standard Go pattern
    ctx := context.Background()
    result, err := client.Parse(ctx, "https://example.com/article")
    if err != nil {
        // Check error type
        if perr, ok := err.(*hermes.ParseError); ok {
            switch perr.Code {
            case hermes.ErrFetch:
                log.Printf("Failed to fetch URL: %v", perr)
            case hermes.ErrSSRF:
                log.Printf("URL blocked by SSRF protection: %v", perr)
            default:
                log.Printf("Parse error: %v", perr)
            }
        }
        log.Fatal(err)
    }
    
    log.Printf("Title: %s", result.Title)
    log.Printf("Author: %s", result.Author)
}
```

### API Server Usage with Semaphore Pattern

```go
func ParseHandler(urls []string) []*hermes.Result {
    sem := make(chan struct{}, 10) // Limit to 10 concurrent
    var wg sync.WaitGroup
    results := make([]*hermes.Result, len(urls))
    
    for i, url := range urls {
        wg.Add(1)
        sem <- struct{}{} // Acquire semaphore
        
        go func(idx int, u string) {
            defer wg.Done()
            defer func() { <-sem }() // Release semaphore
            
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()
            
            if result, err := hermesClient.Parse(ctx, u); err == nil {
                results[idx] = result
            }
        }(i, url)
    }
    
    wg.Wait()
    return results
}
```

### Custom HTTP Client

```go
// Use custom transport with specific settings
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}

client := hermes.New(
    hermes.WithTransport(transport),
    hermes.WithAllowPrivateNetworks(true), // For internal use
)
```

## Success Criteria

- [ ] API can parse single URLs efficiently (caller handles batching)
- [ ] Client can be reused across goroutines safely
- [ ] Context properly threaded throughout
- [ ] Client owns HTTP resources (no globals)
- [ ] Typed errors for programmatic handling
- [ ] Zero dependencies beyond standard library for public API
- [ ] Clean separation between public API and internals
- [ ] 3 lines of code for basic usage
- [ ] All existing tests pass (updated for new API)
- [ ] CLI continues to work with new API
- [ ] Module path remains v1 (`github.com/BumpyClock/hermes`)
- [ ] Standard Go patterns: single input, `(*Result, error)` return

## Notes & Discussions

### Context Support Rationale

Context is properly threaded throughout because:
- Enables proper request cancellation
- Respects caller's timeout requirements
- Supports distributed tracing
- Prevents resource leaks
- Standard in all modern Go libraries

### Concurrency Handling (REVISED)

Caller handles concurrency because:
- Library does ONE thing well: parsing single URLs
- Caller knows their own requirements (rate limits, worker pools, etc.)
- Simpler library implementation and testing
- More composable and flexible
- Follows Unix philosophy

### Error Handling Design (ENHANCED)

Typed `ParseError` with error codes because:
- Enables programmatic error handling
- Distinguishes between error types (network, validation, extraction)
- Supports error wrapping/unwrapping
- Provides context about what failed
- Idiomatic Go error handling

### HTTP Client Ownership

Client-owned HTTP client because:
- No global state
- Proper resource management
- Configurable transport settings
- Connection pooling per client instance
- Testable with custom transports

## Implementation Summary

### Core API (Final Design)

- **3 public methods**: `New()`, `Parse()`, `ParseHTML()`
- **Single responsibility**: Parse one URL at a time
- **Standard Go patterns**: `(*Result, error)` return type
- **Functional options**: Configure client behavior
- **Context-aware**: All parsing methods require `context.Context`
- **Thread-safe**: Client can be shared across goroutines
- **Typed errors**: ParseError with error codes

### What Makes This "Best-in-Class"

1. **Follows 2024-2025 Go conventions** (functional options, context, typed errors)
2. **Unix philosophy** (single responsibility, composable)
3. **Zero unnecessary complexity** (no batch handling, no orchestration)
4. **Idiomatic Go** (interfaces for mocking, proper resource ownership)
5. **Production-ready** (context support, typed errors, configurable security)
6. **3 lines of code** for basic usage

## Next Steps

1. ✅ Create branch `aditya/go-module-refactor`
2. ✅ Update session context with comprehensive plan
3. 🔄 Phase A: Create new public API (wrapping existing)
4. ⏳ Phase B: Thread context through internals
5. ⏳ Phase C: Move pkg to internal
6. ⏳ Phase D: Remove orchestration code
7. ⏳ Phase E: Update CLI
8. ⏳ Phase F: Documentation and examples
9. ⏳ Phase G: Comprehensive testing

**Current Status**: Implementing Phase A - Creating root-level public API