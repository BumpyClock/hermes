# Basic examples

These examples show URL extraction, HTML extraction, and concurrent requests with the Go library.

The first example is a complete program. Later snippets assume that a client and the necessary imports already exist.

## Extract from a URL

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
        hermes.WithUserAgent("ExampleBot/1.0"),
    )

    res, err := client.Parse(context.Background(), "https://www.theguardian.com/technology")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Title:", res.Title)
    fmt.Println("Author:", res.Author)
    fmt.Println("WordCount:", res.WordCount)
}
```

## Extract from HTML

```go
html := `<!doctype html><html><body><article><h1>Title</h1><p>Body...</p></article></body></html>`
res, err := client.ParseHTML(context.Background(), html, "https://example.com/article")
```

## Choose content format

```go
// Markdown content
client := hermes.New(hermes.WithContentType("markdown"))
res, err := client.Parse(context.Background(), "https://example.com/article")
if err != nil {
    log.Fatal(err)
}
fmt.Println(res.Content) // markdown
```

## Concurrency with a semaphore

Import `sync` and `time` for this example. The semaphore limits concurrent requests to five.

```go
urls := []string{"https://example.com/1", "https://example.com/2"}
sem := make(chan struct{}, 5)
var wg sync.WaitGroup
for _, u := range urls {
    wg.Add(1)
    sem <- struct{}{}
    go func(url string) {
        defer wg.Done()
        defer func(){ <-sem }()
        ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
        defer cancel()
        res, err := client.Parse(ctx, url)
        if err == nil {
            fmt.Println(res.Title)
        }
    }(u)
}
wg.Wait()
```
