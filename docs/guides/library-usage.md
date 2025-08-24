# Library Usage

This guide explains how to use Hermes as a Go module in your own projects. It covers
installation, basic parsing, configuration options and high-throughput batch
processing.

## Installation

Add Hermes to your module with `go get`:

```bash
go get github.com/BumpyClock/hermes
```

If you are starting a new project remember to initialize a module first:

```bash
go mod init example.com/myproject
```

## Basic Example

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

    if result.IsError() {
        log.Fatal(result.Message)
    }

    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Content: %s\n", result.Content)
}
```

## Configuring the Parser

`ParserOptions` control extraction behaviour:

```go
opts := &parser.ParserOptions{
    ContentType:   "markdown",             // html, markdown, text or json
    FetchAllPages: true,                    // hint to follow next_page_url
    Headers:       map[string]string{
        "User-Agent": "MyBot/1.0",
    },
}
result, err := p.Parse(url, opts)
```

## Batch Processing

For high volume workloads use `NewHighThroughputParser` which reuses workers and
connections:

```go
urls := []string{
    "https://example.com/1",
    "https://example.com/2",
}

ht := parser.NewHighThroughputParser(&parser.ParserOptions{ContentType: "text"})
results, err := ht.BatchParse(urls, nil)
if err != nil {
    log.Fatal(err)
}

for i, res := range results {
    if res.IsError() {
        log.Printf("failed to parse %s: %s", urls[i], res.Message)
        ht.ReturnResult(res)
        continue
    }
    fmt.Printf("%s -> %d words\n", res.Title, res.WordCount)
    ht.ReturnResult(res)
}
```

## Error Handling and Pagination

Hermes returns a `Result` object even when extraction fails. Always check
`result.IsError()` and the `Message` field before using the content. When a
`next_page_url` is detected you can fetch subsequent pages yourself or use the
`FetchAllPages` option (multi-page merging is still in development).

## Next Steps

- Review the [Parser API](../api/parser.md) for a complete list of options
  and result fields.
- Browse [examples](../examples/basic.md) for more advanced patterns.
- Learn about [custom extractors](custom-extractors.md) to improve extraction
  for specific sites.

