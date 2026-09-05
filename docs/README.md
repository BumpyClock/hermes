# Hermes documentation

Hermes extracts article content and metadata from web pages. Use these guides for the Go library, CLI, and internal extractor system.

## Guides and reference

### Start here

- [Installation](guides/installation.md)
- [Basic usage](guides/basic-usage.md)
- [CLI commands and flags](guides/cli-usage.md)

### API reference

- [Hermes API](api/hermes.md)
- [Configuration](api/configuration.md)
- [Result fields and helpers](api/results.md)
- [Internal extractors](api/extractors.md)

### Architecture

- [Architecture overview](architecture/overview.md)

### Examples

- [Go client examples](examples/basic.md)

## Quick start

```bash
# Install the CLI
go install github.com/BumpyClock/hermes/cmd/hermes@latest

# Parse a URL via CLI
hermes parse https://example.com/article

# Use as a library in Go modules
go get github.com/BumpyClock/hermes
```

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
    client := hermes.New(
        hermes.WithTimeout(30 * time.Second),
        hermes.WithUserAgent("MyApp/1.0"),
    )

    ctx := context.Background()
    result, err := client.Parse(ctx, "https://example.com/article")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Content: %s\n", result.Content)
}
```

## Internal design

The public API exposes the client, options, and results. The [architecture overview](architecture/overview.md) explains the internal parser, extractors, cleaners, and resource package.

## Documentation contributions

1. Describe the reader's task with concrete commands or examples.
2. Check API names and behavior against the source.
3. Test changed examples and local links.
4. Submit changes through the [contribution process](development/contributing.md).
