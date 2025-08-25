# Hermes

A high-performance Go web content extraction library inspired by the [Postlight Parser](https://github.com/postlight/parser). Hermes transforms web pages into clean, structured text with high compatibility with the original JavaScript version while providing significant performance improvements.

## Features

- **Fast Content Extraction**: 2-3x faster than the JavaScript version
- **Memory Efficient**: 50% less memory usage
- **150+ Custom Extractors**: Site-specific parsers for major publications
- **Multiple Output Formats**: HTML, Markdown, plain text, and JSON
- **Pagination Aware**: Detects `next_page_url` for manual multi-page handling
- **CLI Tool**: Command-line interface for single and batch parsing

## Installation

### As a Go Module

```bash
go get github.com/BumpyClock/hermes@latest
```

### CLI Tool

```bash
go install github.com/BumpyClock/hermes/cmd/parser@latest
```

### Build from Source

```bash
git clone https://github.com/BumpyClock/hermes
cd hermes
make build
```

## Usage

### Command Line

```bash
# Parse a URL and output JSON
parser parse https://example.com/article

# Output as markdown
parser parse -f markdown https://example.com/article

# Save to file
parser parse -o article.md -f markdown https://example.com/article

# Multiple URLs with timing
parser parse --timing https://example.com/article1 https://example.com/article2
```

### Go Library

#### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/BumpyClock/hermes"
)

func main() {
    // Create a client with options
    client := hermes.New(
        hermes.WithTimeout(30*time.Second),
        hermes.WithContentType("html"), // "html", "markdown", or "text"
        hermes.WithUserAgent("MyApp/1.0"),
    )
    
    // Parse a URL with context
    ctx := context.Background()
    result, err := client.Parse(ctx, "https://example.com/article")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Author: %s\n", result.Author)
    fmt.Printf("Content: %s\n", result.Content)
    fmt.Printf("Word Count: %d\n", result.WordCount)
}
```

#### Advanced Usage with Custom HTTP Client

```go
package main

import (
    "context"
    "crypto/tls"
    "fmt"
    "net/http"
    "time"
    
    "github.com/BumpyClock/hermes"
)

func main() {
    // Create custom HTTP client with proxy, custom transport, etc.
    customClient := &http.Client{
        Timeout: 60 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: false,
            },
        },
    }
    
    // Create Hermes client with custom HTTP client
    client := hermes.New(
        hermes.WithHTTPClient(customClient),
        hermes.WithContentType("markdown"),
        hermes.WithAllowPrivateNetworks(false), // SSRF protection
    )
    
    // Parse with timeout context
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    result, err := client.Parse(ctx, "https://example.com/article")
    if err != nil {
        if parseErr, ok := err.(*hermes.ParseError); ok {
            fmt.Printf("Parse error [%s]: %v\n", parseErr.Code, parseErr.Err)
        } else {
            log.Fatal(err)
        }
        return
    }
    
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Content: %s\n", result.Content)
}
```

#### Parse Pre-fetched HTML

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/BumpyClock/hermes"
)

func main() {
    client := hermes.New(hermes.WithContentType("text"))
    
    html := `<html><head><title>Test</title></head><body><p>Hello World</p></body></html>`
    
    result, err := client.ParseHTML(context.Background(), html, "https://example.com/test")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Content: %s\n", result.Content)
}
```

## Migration from v0.x to v1.0

If you're upgrading from the old internal API, here are the key changes:

### Old API (v0.x)
```go
import "github.com/BumpyClock/hermes/pkg/parser"

p := parser.New()
result, err := p.Parse(url, &parser.ParserOptions{...})
```

### New API (v1.0+)
```go
import "github.com/BumpyClock/hermes"

client := hermes.New(hermes.WithTimeout(...))
result, err := client.Parse(ctx, url)
```

### Key Changes

1. **Package Import**: Use root package instead of `/pkg/parser`
2. **Context Required**: All methods now require `context.Context` first parameter
3. **Functional Options**: Use `hermes.WithXxx()` options instead of struct fields
4. **Error Types**: New `*hermes.ParseError` type with error codes
5. **HTTP Client**: Client manages its own HTTP client, configurable via options
6. **Content Types**: Set via `WithContentType()` option, affects parser extraction

### Options Mapping

| Old API | New API |
|---------|---------|
| `parser.ParserOptions{ContentType: "markdown"}` | `hermes.WithContentType("markdown")` |
| `parser.ParserOptions{FetchAllPages: true}` | Use `result.NextPageURL` for manual pagination |
| Custom headers in options | Use `hermes.WithHTTPClient()` with custom transport |

## Error Handling

The new API provides structured error handling:

```go
result, err := client.Parse(ctx, url)
if err != nil {
    if parseErr, ok := err.(*hermes.ParseError); ok {
        switch parseErr.Code {
        case hermes.ErrInvalidURL:
            // Handle invalid URL
        case hermes.ErrFetch:
            // Handle fetch error
        case hermes.ErrTimeout:
            // Handle timeout
        case hermes.ErrExtract:
            // Handle extraction error
        default:
            // Handle other errors
        }
    }
}
```

See the [Library Usage Guide](docs/guides/library-usage.md) for more ways to integrate Hermes into your Go projects.

## Development

### Prerequisites

- Go 1.24.6 or later
- Make (optional)

### Setup

```bash
# Clone and setup
git clone https://github.com/BumpyClock/hermes
cd hermes
make dev-setup

# Run tests
make test

# Run with fixtures
make run-fixtures

# Lint code
make lint

# Build binary
make build
```

## Key Dependencies

Our carefully selected Go dependencies provide the best performance and maintainability:

- **goquery**: jQuery-like DOM manipulation (industry standard)
- **html-to-markdown**: HTML to Markdown conversion (v1.6.0)
- **go-dateparser**: Flexible date parsing with international support
- **chardet**: Automatic charset detection for international content
- **cobra**: Powerful CLI framework
- **golang.org/x/text**: Official Go text encoding support

### Testing

The project includes comprehensive unit tests. Compatibility tests with the JavaScript version are planned. The `make test-compatibility` target currently references a non-existent package and will be enabled once the compatibility suite is added.

```bash
# Run all tests
go test ./...

# Test with coverage
go test -cover ./...

# Benchmark tests
make benchmark
```

## Architecture

Hermes follows a modular architecture similar to the JavaScript version:

- **Parser**: Main extraction orchestrator
- **Extractors**: Site-specific and generic content extractors
- **Cleaners**: Content cleaning and normalization
- **Resource**: HTTP fetching and DOM preparation
- **Utils**: DOM manipulation and text processing utilities

## Custom Extractors

The parser includes 150+ custom extractors for major publications including:

- News: NY Times, Washington Post, CNN, The Guardian
- Tech: Ars Technica, The Verge, Wired
- Business: Bloomberg, Reuters
- And many more...

## Performance

Performance varies by site and output format. See benchmark details in `benchmark/README.md`.

Latest benchmark (5 URLs from `benchmark/testurls.txt`):

- JSON output: JS avg 627ms, Go avg 629ms (parity)
- Markdown output: JS avg 173ms, Go avg 652ms (JS faster on this set)

Run the comparison yourself via `benchmark/test-comparison.js` (see docs in `benchmark/README.md`).

Running the bench with 1 url at a time JS comes out slightly faster than go but with twice the memory usage. In API scenarios and processing multiple urls at once GO leaps ahead with approx 20ms per request with around 60mb memory as the efficiency gains of reusing the same HTTP client and goroutines start to show their edge.

## Compatibility

Hermes aims for high compatibility with the JavaScript version:

- Same output formats and extractor definitions
- CLI commands and options are similar
- Next page URL detection is implemented

Note: Use the `next_page_url` field for manual pagination handling when needed.

## TODOs

### Multi-page Article Collection

The multi-page article collection feature is partially implemented but needs integration:

- [ ] **Integration**: Connect `collect_all_pages.go` with main parser pipeline
- [ ] **Configuration**: Wire `FetchAllPages` option to trigger actual multi-page merging
- [ ] **Pipeline**: Implement call to `CollectAllPages` when `NextPageURL` is detected
- [ ] **Testing**: Add comprehensive multi-page extraction tests

**Files requiring work:**

- `pkg/parser/parser.go` - Uncomment and implement `collectAllPages` method
- `pkg/extractors/collect_all_pages.go` - Already implemented, needs integration
- `pkg/parser/extract_all_fields.go` - Add multi-page logic to extraction pipeline

**Current Status:** Next page URL detection works; automatic fetching/merging does not.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original [Postlight Parser](https://github.com/postlight/parser) team
- [goquery](https://github.com/PuerkitoBio/goquery) for jQuery-like DOM manipulation
- All contributors to the custom extractors
