# Hermes API reference

The `github.com/BumpyClock/hermes` package extracts article content and metadata from web pages.

## Contents

- [Client](#client)
- [Options](#options)
- [Parsing](#parsing)
- [Results](#results)
- [Errors](#errors)

## Client

```go
type Client struct { /* ... */ }

func New(opts ...Option) *Client
```

`New` creates a reusable, thread-safe client. Share one client across goroutines.

Example:
```go
client := hermes.New(
    hermes.WithTimeout(30 * time.Second),
    hermes.WithUserAgent("MyApp/1.0"),
)
```

## Options

Hermes uses functional options to configure the client:

```go
type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option
func WithTransport(transport http.RoundTripper) Option
func WithTimeout(timeout time.Duration) Option
func WithUserAgent(userAgent string) Option
func WithAllowPrivateNetworks(allow bool) Option
func WithContentType(contentType string) Option // "html" | "markdown" | "text"
func WithDefinitions(snapshot *Definitions) Option
```

See [Configuration](configuration.md) for option behavior, defaults, and security precautions.
See [Local definitions](definitions.md) for `LoadDefinitions`, immutable snapshots,
schema/capability discovery, diagnostics and the CLI `--definitions` option.

## Parsing

```go
func (c *Client) Parse(ctx context.Context, url string) (*Result, error)
func (c *Client) ParseHTML(ctx context.Context, html, url string) (*Result, error)
```

`Parse` fetches the URL and extracts content. `ParseHTML` extracts content from supplied HTML with the specified base URL.

When extraction uses the last-resort DOM-text fallback, literal text is HTML-escaped
for HTML output and punctuation is backslash-escaped for Markdown output. Plain-text
output preserves literal characters with normalized whitespace. This applies to both
explicit definition snapshots and the legacy parser; text resembling markup is not
promoted into active HTML or Markdown. Plain-text results still require escaping
before insertion into an HTML document.

Examples:
```go
// Parse a URL
ctx := context.Background()
res, err := client.Parse(ctx, "https://example.com/article")
if err != nil { /* handle */ }
fmt.Println(res.Title)

// Parse pre-fetched HTML
html := "<html>...</html>"
res, err := client.ParseHTML(ctx, html, "https://example.com")
```

Use a semaphore to limit concurrent requests:
```go
sem := make(chan struct{}, 10)
var wg sync.WaitGroup
for _, u := range urls {
    wg.Add(1)
    sem <- struct{}{}
    go func(url string) {
        defer wg.Done()
        defer func(){ <-sem }()
        res, err := client.Parse(ctx, url)
        _ = res; _ = err
    }(u)
}
wg.Wait()
```

## Results

See [Results](results.md) for all fields and helper methods.

## Errors

`ParseError` identifies the operation, URL, error code, and underlying error:

```go
type ErrorCode int

const (
    ErrInvalidURL ErrorCode = iota
    ErrFetch
    ErrTimeout
    ErrSSRF
    ErrExtract
    ErrContext
)

type ParseError struct {
    Code ErrorCode
    URL  string
    Op   string // "Parse" or "ParseHTML"
    Err  error  // underlying error (optional)
}
```

Usage:
```go
res, err := client.Parse(ctx, url)
if err != nil {
    var perr *hermes.ParseError
    if errors.As(err, &perr) {
        switch perr.Code {
        case hermes.ErrTimeout:
            // handle timeout
        case hermes.ErrSSRF:
            // handle SSRF protection
        }
    }
}
```
