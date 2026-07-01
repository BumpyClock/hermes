# Project Overview

Hermes is a high-performance Go web content extraction library inspired by Postlight Parser that transforms web pages into clean, structured data. It extracts article content, titles, authors, dates, images, and more from any URL using site-specific custom parsers and generic fallback extraction. This Go implementation provides 2-3x performance improvements over the JavaScript version while maintaining compatibility.

## Development Commands

### Build Commands

- `make build` - Build the CLI binary to `bin/hermes`
- `go build ./cmd/hermes` - Build CLI directly with Go
- `make install` - Install CLI tool globally
- `go install github.com/BumpyClock/hermes/cmd/hermes@latest` - Install from remote

### Testing Commands

- `make test` - Run all tests with coverage
- `go test -v ./...` - Run all tests verbosely
- `go test -cover ./...` - Run tests with coverage report
- `make benchmark` - Run performance benchmarks
- `go test -bench=. -benchmem ./...` - Detailed benchmark with memory stats
- `make test-compatibility` - Run JavaScript compatibility tests (when implemented)

### CLI Usage

- `./bin/hermes parse https://example.com/article` - Parse URL to JSON
- `./bin/hermes parse -f markdown https://example.com` - Output as Markdown
- `./bin/hermes parse -o output.md -f markdown https://example.com` - Save to file
- `./bin/hermes parse --timing https://example.com` - Show parsing timing

### Development

- `make deps` - Download and tidy dependencies
- `make clean` - Clean build artifacts and test cache
- `make lint` - Run golangci-lint (requires golangci-lint installation)
- `make dev-setup` - Setup development environment with fixtures

## Go Architecture Overview

### Core Components

**Root Package (**`github.com/BumpyClock/hermes`)

- `client.go` - Main Client struct with Parse methods and options
- `parser.go` - Core parsing orchestration and pipeline
- `options.go` - Functional options pattern (WithTimeout, WithContentType, etc.)
- `result.go` - ParseResult struct defining output format
- `errors.go` - Structured error handling with error codes

**Internal Packages**

- `internal/extractors/` - Custom and generic content extractors
- `internal/cleaners/` - Field-specific content cleaning and normalization
- `internal/utils/` - DOM manipulation and text processing utilities

**CLI (**`cmd/hermes/`)

- Cobra-based command line interface with format options and batch processing

### Custom Extractor System

The Go implementation uses a sophisticated extractor architecture:

**Registry System (**`internal/extractors/custom/index.go`)

- Cached extractor inventory built from `GetAllCustomExtractors`
- Domain-to-extractor mapping with support for multiple domains per extractor
- Deterministic conflict handling for overlapping domains

**Extractor Interface (**`internal/extractors/custom/extractor_interface.go`)

```go
type CustomExtractor struct {
    Domain           string
    SupportedDomains []string
    Title            *FieldExtractor
    Author           *FieldExtractor  
    Content          *ContentExtractor
    DatePublished    *FieldExtractor
    LeadImageURL     *FieldExtractor
    // ... other fields
}

type ContentExtractor struct {
    *FieldExtractor
    Clean      []string                    // CSS selectors to remove
    Transforms map[string]TransformFunction // Element transformations
}
```

**Transform Functions**

- `StringTransform` - Simple tag name changes (e.g., `noscript` → `div`)
- `FunctionTransform` - Custom DOM manipulation with goquery Selection

### 150+ Custom Extractors

All extractors are located in `internal/extractors/custom/` and registered in `index.go`:

```go
func GetAllCustomExtractors() map[string]*CustomExtractor {
    return map[string]*CustomExtractor{
        "MediumExtractor": GetMediumExtractor(),
        "NYTimesExtractor": GetNYTimesExtractor(), 
        "DaringFireballExtractor": GetDaringFireballExtractor(),
        // ... 150+ more extractors
    }
}
```

**Categories include:**

- Content Platforms: Medium, Reddit, Twitter, Wikipedia
- News Sites: NY Times, Washington Post, CNN, The Guardian
- Tech Publications: Ars Technica, The Verge, Wired
- International Sites: Japanese tech sites, European publications
- Blog & Commentary: Daring Fireball, personal blogs

## Key Development Workflows

### Adding a Custom Extractor

1. **Create extractor file**: `internal/extractors/custom/domain_com.go`
2. **Implement CustomExtractor struct** with selectors and cleaning rules
3. **Add getter function**: `GetDomainExtractor() *CustomExtractor`
4. **Register in index.go**: Add to `GetAllCustomExtractors()` map
5. **Test extraction**: Use CLI to verify output quality

### Extractor Development Pattern

```go
package custom

import (
    "strings"
    "github.com/PuerkitoBio/goquery"
)

var SiteExtractor = &CustomExtractor{
    Domain: "example.com", 
    SupportedDomains: []string{"www.example.com"},
    
    Title: &FieldExtractor{
        Selectors: []interface{}{
            "h1.headline",
            "title", // fallback
        },
    },
    
    Content: &ContentExtractor{
        FieldExtractor: &FieldExtractor{
            Selectors: []interface{}{"article", ".content"},
        },
        Clean: []string{".ads", "nav", "footer"},
        Transforms: map[string]TransformFunction{
            "img": &FunctionTransform{
                Fn: func(sel *goquery.Selection) error {
                    // Custom image handling
                    return nil
                },
            },
        },
    },
}

func GetSiteExtractor() *CustomExtractor {
    return SiteExtractor
}
```

### Testing Custom Extractors

- **Unit Tests**: Create `*_test.go` files with fixture-based testing
- **Integration Tests**: Use real URLs to verify extraction quality
- **CLI Testing**: `./bin/hermes parse https://site.com/article -f markdown`
- **Benchmark Testing**: Add to benchmark suite for performance tracking

## Content Processing Pipeline

1. **URL Fetch** - HTTP client with retry logic and encoding detection
2. **DOM Parse** - goquery-based HTML parsing and normalization
3. **Extractor Selection** - Domain matching or HTML-based detection
4. **Field Extraction** - Title, author, content, date, images using CSS selectors
5. **Content Cleaning** - Remove ads, navigation, unwanted elements
6. **Transform Application** - Custom DOM transformations per site
7. **Output Formatting** - HTML, Markdown, or plain text conversion
8. **Result Assembly** - Structured ParseResult with metadata

## Key Dependencies

- **goquery** - jQuery-like DOM manipulation (industry standard)
- **html-to-markdown** - HTML to Markdown conversion with configurable rules
- **go-dateparser** - Flexible date parsing with international support
- **chardet** - Automatic charset detection for international content
- **cobra** - CLI framework with subcommands and flag parsing
- **bluemonday** - HTML sanitization for security
- **golang.org/x/text** - Text encoding and normalization

## Performance Characteristics

- **2-3x faster** than JavaScript version for single URL parsing
- **50% less memory** usage compared to Node.js implementation
- **Excellent concurrency** - HTTP client reuse and goroutines for batch processing
- **\~20ms per request** in API scenarios with connection pooling
- **\~60MB memory** footprint for sustained high-throughput processing

## Error Handling

Structured error types with specific error codes:

```go
if parseErr, ok := err.(*hermes.ParseError); ok {
    switch parseErr.Code {
    case hermes.ErrInvalidURL:   // URL validation failed
    case hermes.ErrFetch:        // HTTP fetch failed  
    case hermes.ErrTimeout:      // Request timeout
    case hermes.ErrExtract:      // Content extraction failed
    }
}
```

## File Organization

- `cmd/hermes/` - CLI application entry point
- `examples/` - Usage examples (basic, concurrent, API server)
- `internal/extractors/custom/` - 150+ site-specific extractors
- `internal/extractors/` - Generic extraction algorithms and pipeline
- `internal/cleaners/` - Field-specific content cleaning functions
- `internal/utils/` - Text processing and DOM utilities
- `benchmark/` - Performance testing and comparison tools

## Development Notes

### Multi-page Article Support

- **Detection**: `NextPageURL` field populated when pagination is detected
- **Manual Handling**: Use `result.NextPageURL` for recursive fetching
- **Automatic Collection**: Feature partially implemented in `internal/extractors/collect_all_pages.go`

### Compatibility with JavaScript Version

- Same output structure and field names
- Compatible extractor selector syntax
- Similar CLI command structure and options
- Cross-platform testing framework (planned)

### Memory and Performance Optimization

- Lazy loading of custom extractors via factory pattern
- HTTP client connection pooling and reuse
- Efficient DOM traversal with goquery
- Minimal memory allocations in hot paths
