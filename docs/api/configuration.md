# Configuration API Reference

This document covers available configuration for the public Hermes client.

Hermes is configured via functional options passed to `hermes.New(...)`. There is no public `ParserOptions` type, no environment-variable loader, and no YAML/JSON config reader in the library.

## Table of Contents

- [Client Options](#client-options)
- [Content Type](#content-type)
- [HTTP/TLS](#httptls)
- [Security](#security)
- [Examples](#examples)

## Client Options

```go
func WithHTTPClient(httpClient *http.Client) Option
func WithTransport(transport http.RoundTripper) Option
func WithTimeout(timeout time.Duration) Option
func WithUserAgent(userAgent string) Option
func WithAllowPrivateNetworks(allow bool) Option
func WithContentType(contentType string) Option // "html" | "markdown" | "text"
```

- WithHTTPClient: Provide a fully configured `*http.Client` (proxies, TLS, pools, timeouts).
- WithTransport: Provide a custom `http.RoundTripper` if you are not supplying a full client.
- WithTimeout: Set request timeout on Hermes-created clients, or on a custom client when explicitly composed with `WithHTTPClient`.
- WithUserAgent: Override the default `User-Agent` string.
- WithAllowPrivateNetworks: Permit private network/localhost URLs (defaults to false for SSRF protection).
- WithContentType: Select output format for the `Result.Content` field.

## Content Type

`WithContentType` controls the format of the extracted `Result.Content`:

- `"html"` (default): Clean, sanitized HTML.
- `"markdown"`: Markdown conversion of the content.
- `"text"`: Plain text.

This option affects only the content body. All metadata fields are structured the same regardless of content type.

## HTTP/TLS

- Prefer `WithHTTPClient` to supply your own client if you need proxies, retry logic, or custom TLS settings.
- Use `WithTransport` to swap only the transport on the internally created client.
- `WithUserAgent` sets the `User-Agent` header; other headers are not currently configurable at the client level.

## Security

- `WithAllowPrivateNetworks(false)` (default) blocks private IP ranges and localhost to mitigate SSRF.
- Set `WithAllowPrivateNetworks(true)` only in trusted environments when you need to parse internal URLs.

## Examples

Basic configuration:

```go
client := hermes.New(
    hermes.WithTimeout(20*time.Second),
    hermes.WithUserAgent("MyApp/2.0"),
    hermes.WithContentType("markdown"),
)
```

Custom HTTP client with proxy and TLS:

```go
tr := &http.Transport{ /* custom TLS, proxy, pools */ }
httpClient := &http.Client{ Transport: tr, Timeout: 45*time.Second }

client := hermes.New(
    hermes.WithHTTPClient(httpClient),
)
```

Allow private networks (in a trusted, internal tool):

```go
client := hermes.New(
    hermes.WithAllowPrivateNetworks(true),
)
```
