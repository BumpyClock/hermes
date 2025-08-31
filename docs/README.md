# Hermes Documentation

Welcome to the Hermes documentation. Hermes is a high-performance Go library and CLI for extracting clean, structured content from web pages.

## Table of Contents

### Getting Started

- [Installation & Setup](guides/installation.md) - Quick start guide and installation instructions
- [Basic Usage](guides/basic-usage.md) - First steps with the Go API and CLI
- [CLI Usage](guides/cli-usage.md) - Command line interface documentation

### API Reference

- [Hermes API](api/hermes.md) - Public Go API (client, options, errors)
- [Configuration](api/configuration.md) - Client options and behaviors
- [Results](api/results.md) - Result fields and helpers

### Architecture & Design

- [Architecture Overview](architecture/overview.md) - System design and components

### Examples

- [Basic Examples](examples/basic.md) - Practical usage examples with the Go client

## Quick Start

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

## Architecture at a Glance

Hermes uses a modular internal architecture with a small public surface (client, options, results). See [Architecture Overview](architecture/overview.md) for details on extractors, cleaners, and the resource layer.

## Documentation Standards

All documentation follows these principles:

- **Practical examples** for every feature
- **Complete API coverage** with parameters and return values
- **Architecture explanations** with diagrams where helpful
- **Performance considerations** for production usage
- **Migration guides** from other parsers

## Contributing to Documentation

We welcome contributions to improve our documentation:

1. Fork the repository
2. Create a feature branch for your documentation changes
3. Follow the documentation style guide
4. Test code examples
5. Submit a pull request

See [Contributing Guide](development/contributing.md) for detailed instructions.
