# Basic Usage Guide

This guide covers fundamental usage patterns for Hermes with both the CLI and the Go API.

## Table of Contents

- [CLI](#cli)
- [Library](#library)
- [Content Formats](#content-formats)
- [Error Handling](#error-handling)
- [Concurrency](#concurrency)

## CLI

Install the CLI:

```bash
go install github.com/BumpyClock/hermes/cmd/hermes@latest
```

Parse a single URL (JSON by default):

```bash
hermes parse https://example.com/article
```

Select output format:

```bash
hermes parse -f markdown https://example.com/article
hermes parse -f html https://example.com/article
hermes parse -f text https://example.com/article
```

Multiple URLs (JSON array):

```bash
hermes parse https://example.com/1 https://example.com/2
```

See [CLI Usage](cli-usage.md) for all flags and details.

## Library

Add the module and import:

```bash
go get github.com/BumpyClock/hermes
```

Basic usage:

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
        hermes.WithTimeout(30*time.Second),
        hermes.WithUserAgent("ExampleApp/1.0"),
        hermes.WithContentType("html"),
    )

    ctx := context.Background()
    res, err := client.Parse(ctx, "https://example.com/article")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(res.Title)
}
```

Parse pre-fetched HTML:

```go
html := "<html>...</html>"
res, err := client.ParseHTML(context.Background(), html, "https://example.com")
```

## Content Formats

Control the content format via `WithContentType`:

- `"html"` (default): cleans and returns HTML
- `"markdown"`: converts content to Markdown
- `"text"`: returns plain text

Export complete result as Markdown:

```go
md := res.FormatMarkdown()
fmt.Println(md)
```

## Error Handling

Typed errors make it easy to branch on failure kinds:

```go
res, err := client.Parse(ctx, url)
if err != nil {
    var perr *hermes.ParseError
    if errors.As(err, &perr) {
        switch perr.Code {
        case hermes.ErrInvalidURL:
            // fix URL
        case hermes.ErrTimeout:
            // retry or increase timeout
        case hermes.ErrSSRF:
            // blocked by private network protection
        }
    }
}
```

## Concurrency

Hermes parses one URL at a time per call. For concurrency, fan out requests with a semaphore to bound parallelism:

```go
sem := make(chan struct{}, 10)
var wg sync.WaitGroup
for _, u := range urls {
    wg.Add(1)
    sem <- struct{}{}
    go func(url string) {
        defer wg.Done()
        defer func(){ <-sem }()
        ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
        defer cancel()
        _, _ = client.Parse(ctx, url)
    }(u)
}
wg.Wait()
```
