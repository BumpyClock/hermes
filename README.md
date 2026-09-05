# Hermes

Hermes is a Go library and CLI that extracts article content and metadata from web pages. It uses site-specific extractors with a generic fallback, based on the [Postlight Parser](https://github.com/postlight/parser).

## Features

- Extract article text, titles, authors, dates, and images.
- Use site-specific rules for supported sites and generic extraction for other pages.
- Return content as HTML, Markdown, or plain text. The CLI also supports JSON output.
- Parse one URL or a batch with the CLI.

## Installation

See [Installation](docs/guides/installation.md) for module, CLI, and source build instructions.

## Usage

### Command line

```bash
# Parse a URL and output JSON
hermes parse https://example.com/article

# Output as markdown
hermes parse -f markdown https://example.com/article

# Save to file
hermes parse -o article.md -f markdown https://example.com/article

# Multiple URLs with timing
hermes parse --timing https://example.com/article1 https://example.com/article2
```

### Go library

#### Basic usage

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

#### Custom HTTP client

```go
package main

import (
    "context"
    "crypto/tls"
    "fmt"
    "log"
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
        hermes.WithAllowPrivateNetworks(false), // Validate the initial URL for private addresses.
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

See [security limitations](docs/api/configuration.md#security) before you accept untrusted URLs.

#### Parse HTML from a string

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

For code that uses the old internal APIs, replace the imports with the public root package.

### Public API

```go
import "github.com/BumpyClock/hermes"

client := hermes.New()
result, err := client.Parse(ctx, url)
```

### API changes

- The public package is `github.com/BumpyClock/hermes`.
- Both parse methods require `context.Context` as the first argument.
- Functional options replace configuration struct fields.
- Parse errors use `*hermes.ParseError` and an error code.
- Each client has an HTTP client, which options can replace or configure.
- `WithContentType()` selects the extracted content format.

### Option equivalents

| Old API | New API |
|---------|---------|
| `parser.ParserOptions{ContentType: "markdown"}` | `hermes.WithContentType("markdown")` |
| `parser.ParserOptions{FetchAllPages: true}` | No public equivalent. `hermes.Result` does not expose `NextPageURL`. |
| Custom headers in options | Use `hermes.WithHTTPClient()` with custom transport |

## Parse errors

Use `ParseError.Code` to distinguish failure types:

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

## Development

### Prerequisites

- The Go version in [go.mod](go.mod)
- Make for the `make` commands below

### Setup

```bash
# Clone and setup
git clone https://github.com/BumpyClock/hermes
cd hermes
make deps

# Run tests
make test

# Lint code
make lint

# Build binary
make build
```

## Dependencies

- `goquery` provides DOM queries and manipulation.
- `html-to-markdown` converts HTML to Markdown.
- `go-dateparser` parses dates.
- `chardet` detects character encodings.
- `cobra` defines CLI commands and flags.
- `golang.org/x/text` converts text encodings.

See [go.mod](go.mod) for dependency versions.

### Tests

The repository has unit tests and benchmarks. The [comparison scripts](benchmark/README.md) compare CLI output with the JavaScript parser.

```bash
# Run all tests
go test ./...

# Test with coverage
go test -cover ./...

# Benchmark tests
make benchmark
```

## Architecture

The public client calls the internal parser. The parser fetches HTML, selects an extractor, cleans content, and assembles a result.

See the [architecture overview](docs/architecture/overview.md) for package responsibilities.

## Custom extractors

The [custom extractor registry](internal/extractors/custom/index.go) includes rules for these sites and others:

- News: NY Times, Washington Post, CNN, The Guardian
- Tech: Ars Technica, The Verge, Wired
- Business: Bloomberg, Reuters

## Performance

Performance depends on the site, network, output format, and client reuse. The repository does not establish a general speed or memory advantage over JavaScript.

Use the [benchmark instructions](benchmark/README.md) to compare results for your workload.

## Compatibility

Hermes uses extractor rules based on the JavaScript parser, but output can differ. The comparison scripts help identify those differences.

The internal parser detects next-page URLs. The public Go `Result` and CLI do not expose `next_page_url` or automatic page collection.

## Contributing

See the [contribution guide](docs/development/contributing.md) for source setup, tests, and pull request requirements.

## License

Hermes uses the [MIT License](LICENSE).

## Acknowledgments

- Original [Postlight Parser](https://github.com/postlight/parser) team
- [goquery](https://github.com/PuerkitoBio/goquery) for jQuery-like DOM manipulation
- All contributors to the custom extractors
