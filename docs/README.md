# Hermes Documentation

Welcome to the comprehensive documentation for Hermes, a high-performance Go web content extraction library.

## 📖 Table of Contents

### Getting Started
- [Installation & Setup](guides/installation.md) - Quick start guide and installation instructions
- [Basic Usage](guides/basic-usage.md) - Your first steps with Hermes
- [CLI Usage](guides/cli-usage.md) - Command line interface documentation

### API Reference
- [Parser API](api/parser.md) - Core parser interface and methods
- [Extractors](api/extractors.md) - Custom and generic extractors
- [Configuration](api/configuration.md) - Parser options and settings
- [Results](api/results.md) - Result structures and formatting

### Architecture & Design
- [Architecture Overview](architecture/overview.md) - System design and components

### Guides & Tutorials
- [CLI Usage](guides/cli-usage.md) - Detailed CLI commands and flags
- [Basic Usage](guides/basic-usage.md) - Common patterns and examples

### Development
- See repository README for development setup, testing, and build commands

### Examples
- [Basic Examples](examples/basic.md) - Practical usage examples

## 🚀 Quick Start

```bash
# Install Hermes
go install github.com/BumpyClock/hermes/cmd/parser@latest

# Parse a URL
parser parse https://example.com/article

# Use as library
go get github.com/BumpyClock/hermes
```

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
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Content: %s\n", result.Content)
}
```

## 🏗️ Architecture at a Glance

Hermes is built with a modular architecture:

- **Parser**: Main orchestrator for content extraction
- **Extractors**: Site-specific and generic content extractors  
- **Cleaners**: Content cleaning and normalization
- **Resource Layer**: HTTP fetching and DOM preparation
- **Utils**: DOM manipulation and text processing utilities

## 📊 Performance

- **2-3x faster** than JavaScript implementations
- **50% less memory** usage
- **150+ custom extractors** for major publications
- **Multiple output formats** (HTML, Markdown, JSON, Text)

## 📝 Documentation Standards

All documentation follows these principles:
- **Practical examples** for every feature
- **Complete API coverage** with parameters and return values
- **Architecture explanations** with diagrams where helpful
- **Performance considerations** for production usage
- **Migration guides** from other parsers

## 🤝 Contributing to Documentation

We welcome contributions to improve our documentation:

1. Fork the repository
2. Create a feature branch for your documentation changes
3. Follow the documentation style guide
4. Test code examples
5. Submit a pull request

See [Contributing Guide](development/contributing.md) for detailed instructions.
