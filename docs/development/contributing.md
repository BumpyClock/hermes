# Contributing Guide

Thank you for your interest in contributing to Hermes! This guide covers everything you need to know to contribute effectively.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Process](#development-process)
- [Code Standards](#code-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)

## Getting Started

### Prerequisites

Before contributing, ensure you have:

1. **Go 1.24.6 or later** installed
2. **Git** for version control
3. **Make** (optional, for convenience commands)
4. **golangci-lint** for code quality checks

### Setting Up Your Development Environment

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/hermes.git
   cd hermes
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/BumpyClock/hermes.git
   ```

4. **Install dependencies**:
   ```bash
   make deps
   ```

5. **Verify setup**:
   ```bash
   make test
   make lint
   ```

## Development Process

### Workflow Overview

1. **Create an issue** (for significant changes)
2. **Create a feature branch**
3. **Make your changes**
4. **Write/update tests**
5. **Update documentation**
6. **Submit a pull request**

### Branch Naming

Use descriptive branch names following this pattern:

```bash
# Features
feature/add-custom-extractor-registry
feature/batch-processing-improvements

# Bug fixes
fix/memory-leak-in-object-pools
fix/unicode-handling-edge-case

# Documentation
docs/api-reference-updates
docs/contributing-guide-improvements

# Refactoring
refactor/simplify-extractor-interface
refactor/optimize-dom-processing
```

### Commit Message Format

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

**Examples:**
```bash
feat(parser): add support for custom timeout configuration

fix(extractors): handle malformed JSON in custom extractor definitions

docs(api): update parser configuration examples with new options

perf(pools): optimize object allocation in result pools

test(integration): add tests for multi-page extraction scenarios
```

## Code Standards

### Go Style Guidelines

Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html).

#### Package Organization

```go
// Package structure
package parser

import (
    // Standard library imports first
    "context"
    "fmt"
    "strings"
    
    // Third-party imports second
    "github.com/PuerkitoBio/goquery"
    "github.com/spf13/cobra"
    
    // Local imports last
    "github.com/BumpyClock/hermes/pkg/cleaners"
    "github.com/BumpyClock/hermes/pkg/utils/dom"
)
```

#### Function Documentation

All exported functions must have documentation comments:

```go
// Parse extracts content from the specified URL using the configured options.
// It returns a Result containing the extracted content and metadata, or an error
// if the URL cannot be fetched or parsed.
//
// The opts parameter is optional; if nil, default options will be used.
// Custom extractors take precedence over generic extraction when available.
//
// Example:
//   result, err := parser.Parse("https://example.com", &ParserOptions{
//       ContentType: "markdown",
//   })
func (p *Parser) Parse(url string, opts *ParserOptions) (*Result, error) {
    // implementation
}
```

#### Error Handling

Use descriptive error messages with context:

```go
// Good
return nil, fmt.Errorf("failed to parse URL %s: %w", url, err)

// Good - structured error with context
return nil, &ParseError{
    URL: url,
    Operation: "content_extraction",
    Err: err,
}

// Avoid - generic error without context
return nil, err
```

#### Variable Naming

Use clear, descriptive names:

```go
// Good
extractorRegistry := custom.NewExtractorRegistry()
contentCleaner := cleaners.NewContentCleaner()

// Avoid - unclear abbreviations
reg := custom.NewExtractorRegistry()
cc := cleaners.NewContentCleaner()
```

### Code Formatting

Use `gofumpt` for consistent formatting:

```bash
# Format all code
gofumpt -w .

# Check formatting
gofumpt -d .
```

### Linting

Run `golangci-lint` to catch common issues:

```bash
# Run linter
make lint

# Or manually
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix
```

## Testing Guidelines

### Test Structure

Follow the standard Go testing conventions:

```go
func TestParser_Parse(t *testing.T) {
    tests := []struct {
        name        string
        url         string
        options     *ParserOptions
        want        *Result
        wantErr     bool
        wantErrType error
    }{
        {
            name: "successful extraction with default options",
            url:  "https://example.com/article",
            options: nil,
            want: &Result{
                Title: "Example Article",
                WordCount: 150,
            },
            wantErr: false,
        },
        {
            name: "invalid URL returns error",
            url:  "not-a-valid-url",
            options: nil,
            want: nil,
            wantErr: true,
            wantErrType: &url.Error{},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := New()
            got, err := p.Parse(tt.url, tt.options)
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.wantErrType != nil {
                    assert.IsType(t, tt.wantErrType, err)
                }
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.want.Title, got.Title)
            assert.Equal(t, tt.want.WordCount, got.WordCount)
        })
    }
}
```

### Test Categories

#### Unit Tests
- Test individual functions and methods
- Use mocks for external dependencies
- Fast execution (< 100ms per test)

```go
func TestExtractor_ExtractTitle(t *testing.T) {
    html := `<html><head><title>Test Title</title></head></html>`
    doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
    
    extractor := &GenericExtractor{}
    title := extractor.ExtractTitle(doc, "https://example.com")
    
    assert.Equal(t, "Test Title", title)
}
```

#### Integration Tests
- Test component interactions
- Use real HTTP requests (with fixtures when possible)
- Tagged with `// +build integration`

```go
//go:build integration
// +build integration

func TestParser_RealWorldExtraction(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    p := New()
    result, err := p.Parse("https://www.theguardian.com/technology", nil)
    
    assert.NoError(t, err)
    assert.False(t, result.IsError())
    assert.NotEmpty(t, result.Title)
    assert.Greater(t, result.WordCount, 100)
}
```

#### Benchmark Tests
- Measure performance characteristics
- Include memory allocation metrics

```go
func BenchmarkParser_Parse(b *testing.B) {
    p := New()
    url := "https://httpbin.org/html"
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        result, err := p.Parse(url, nil)
        if err != nil {
            b.Fatal(err)
        }
        if result.IsError() {
            b.Fatal(result.Message)
        }
    }
}
```

### Test Requirements

For all contributions:

1. **Unit tests** for new functions and methods
2. **Integration tests** for new features
3. **Benchmark tests** for performance-critical code
4. **Test coverage** should not decrease
5. **All tests must pass** before merging

Run tests:

```bash
# All tests
make test

# Unit tests only
go test ./... -short

# Integration tests
go test ./... -tags=integration

# Benchmarks
go test -bench=. ./...

# Coverage
go test -cover ./...
```

## Documentation

### Code Documentation

1. **Package documentation** in `doc.go` files
2. **Function documentation** for all exported functions
3. **Type documentation** for all exported types
4. **Example functions** for complex APIs

```go
// Package parser provides high-performance web content extraction capabilities.
//
// The parser package implements a content extraction system inspired by
// Postlight Parser, offering both site-specific custom extractors and
// generic fallback extraction algorithms.
//
// Basic usage:
//   p := parser.New()
//   result, err := p.Parse("https://example.com/article", nil)
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Println(result.Title)
package parser
```

### API Documentation

Update relevant documentation files in `docs/`:

- `docs/api/` - API reference documentation
- `docs/guides/` - User guides and tutorials
- `docs/examples/` - Code examples
- `docs/architecture/` - Architecture documentation

### Example Code

Provide working examples for new features:

```go
func ExampleParser_Parse() {
    p := parser.New()
    
    result, err := p.Parse("https://example.com/article", &parser.ParserOptions{
        ContentType: "markdown",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    if result.IsError() {
        log.Fatal(result.Message)
    }
    
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Word Count: %d\n", result.WordCount)
    
    // Output:
    // Title: Example Article
    // Word Count: 250
}
```

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run quality checks**:
   ```bash
   make test
   make lint
   make benchmark  # For performance changes
   ```

3. **Update documentation** if needed

4. **Squash commits** if multiple commits address the same logical change

### PR Description Template

Use this template for your pull request description:

```markdown
## Description
Brief description of the changes made.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)
- [ ] Documentation update

## Changes Made
- List of specific changes
- Use bullet points for clarity

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] All tests pass
- [ ] Benchmarks run (for performance changes)

## Documentation
- [ ] Code comments updated
- [ ] API documentation updated
- [ ] User guide updated (if applicable)

## Checklist
- [ ] Code follows the project's style guidelines
- [ ] Self-review of the code completed
- [ ] Tests added for new functionality
- [ ] All tests pass
- [ ] Documentation updated
- [ ] No breaking changes (or breaking changes documented)

## Related Issues
Closes #<issue_number>
Related to #<issue_number>
```

### Review Process

1. **Automated checks** must pass (CI/CD)
2. **Code review** by maintainers
3. **Testing** on multiple platforms
4. **Documentation review**
5. **Approval** and merge

## Issue Guidelines

### Reporting Bugs

Use this template for bug reports:

```markdown
## Bug Description
A clear and concise description of what the bug is.

## To Reproduce
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

## Expected Behavior
A clear description of what you expected to happen.

## Actual Behavior
What actually happened.

## Environment
- OS: [e.g. macOS 12.0]
- Go version: [e.g. 1.24.6]
- Hermes version: [e.g. v0.1.0]

## Additional Context
Add any other context about the problem here.

## Minimal Reproduction
```go
// Minimal code example that reproduces the issue
package main

import "github.com/BumpyClock/hermes/pkg/parser"

func main() {
    p := parser.New()
    // Code that demonstrates the bug
}
```
```

### Feature Requests

Use this template for feature requests:

```markdown
## Feature Description
A clear and concise description of the feature you'd like to see.

## Problem Statement
What problem does this feature solve? What's the current limitation?

## Proposed Solution
Describe the solution you'd like to see implemented.

## Alternatives Considered
Describe any alternative solutions or features you've considered.

## Use Case
Provide a concrete example of how this feature would be used.

## Implementation Notes
Any thoughts on how this might be implemented (optional).

## Additional Context
Add any other context or screenshots about the feature request here.
```

### Getting Help

For questions and support:

1. **Check existing documentation** first
2. **Search existing issues** for similar questions
3. **Ask in discussions** for general questions
4. **Create an issue** for specific bugs or feature requests

## Contributing Areas

We welcome contributions in these areas:

### High Priority
- **Custom extractors** for new websites
- **Performance optimizations**
- **Bug fixes**
- **Test coverage improvements**

### Medium Priority
- **Documentation improvements**
- **Example code**
- **Build and CI improvements**
- **Code refactoring**

### Future Features
- **Plugin system** for custom extractors
- **Configuration management**
- **Advanced caching**
- **Monitoring and metrics**

## Community Guidelines

### Code of Conduct

We follow the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

### Communication

- **Be respectful** and constructive in all interactions
- **Ask questions** when you're unsure
- **Provide context** when reporting issues
- **Be patient** with review processes
- **Help others** when you can

### Recognition

Contributors are recognized in:
- **CONTRIBUTORS.md** file
- **Release notes** for significant contributions
- **GitHub contributors** page

## Getting Help

If you need help contributing:

1. **Read this guide** thoroughly
2. **Check the documentation** in `docs/`
3. **Look at existing code** for patterns
4. **Ask questions** in GitHub Discussions
5. **Join community channels** (if available)

Thank you for contributing to Hermes! Your contributions help make web content extraction better for everyone.