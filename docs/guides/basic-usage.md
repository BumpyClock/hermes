# Basic usage

Use Hermes through the CLI or the Go API to extract article content and metadata.

## Contents

- [CLI](#cli)
- [Library](#library)
- [Content formats](#content-formats)
- [Error handling](#error-handling)
- [Concurrency](#concurrency)

## CLI

Install the CLI:

```bash
go install github.com/BumpyClock/hermes/cmd/hermes@latest
```

Parse one URL. The default output format is JSON.

```bash
hermes parse https://example.com/article
```

Select the output format:

```bash
hermes parse -f markdown https://example.com/article
hermes parse -f html https://example.com/article
hermes parse -f text https://example.com/article
```

Parse multiple URLs to produce a JSON array:

```bash
hermes parse https://example.com/1 https://example.com/2
```

See [CLI usage](cli-usage.md) for flags and output rules.

## Library

Add the module:

```bash
go get github.com/BumpyClock/hermes
```

Create a client and parse a URL:

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

Parse HTML that you already have:

```go
html := "<html>...</html>"
res, err := client.ParseHTML(context.Background(), html, "https://example.com")
```

## Content formats

Select the content format with `WithContentType`:

- `"html"` returns clean HTML. This is the default.
- `"markdown"` converts content to Markdown.
- `"text"` returns plain text.

Format the result as Markdown with a metadata header:

`FormatMarkdown` preserves `res.Content` as supplied. Select `WithContentType("markdown")` before the parse call to obtain Markdown content.

```go
md := res.FormatMarkdown()
fmt.Println(md)
```

## Error handling

Import `errors` to inspect a `ParseError` with `errors.As`:

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

Each `Parse` call handles one URL. A semaphore limits the number of concurrent calls.

Import `sync` and `time` for this example:

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
