# Configuration API reference

Pass functional options to `hermes.New(...)` to configure the client. The library has no public `ParserOptions` type, environment-variable loader, or YAML/JSON configuration reader.

## Contents

- [Client options](#client-options)
- [Content type](#content-type)
- [HTTP/TLS](#httptls)
- [Security](#security)
- [Examples](#examples)

## Client options

```go
func WithHTTPClient(httpClient *http.Client) Option
func WithTransport(transport http.RoundTripper) Option
func WithTimeout(timeout time.Duration) Option
func WithUserAgent(userAgent string) Option
func WithAllowPrivateNetworks(allow bool) Option
func WithContentType(contentType string) Option // "html" | "markdown" | "text"
```

| Option | Behavior |
| --- | --- |
| `WithHTTPClient` | Uses a supplied `*http.Client` for proxies, TLS, connection pools, and timeouts. |
| `WithTransport` | Sets the HTTP transport. |
| `WithTimeout` | Sets the request timeout. Hermes defaults to 30 seconds and preserves a supplied client's timeout unless this option is explicit. |
| `WithUserAgent` | Sets the `User-Agent` header. The default is `hermes.DefaultUserAgent`. |
| `WithAllowPrivateNetworks` | Permits private network and localhost URLs. The default is `false`. |
| `WithContentType` | Sets the format of `Result.Content`. The default is `"html"`. |

## Content type

`WithContentType` controls the format of the extracted `Result.Content`:

- `"html"`: Clean, sanitized HTML. This is the default.
- `"markdown"`: Markdown conversion of the content.
- `"text"`: Plain text.

This option affects only the content body. Metadata fields keep the same structure for all content types.

## HTTP/TLS

- Use `WithHTTPClient` when you need proxies, retry logic, or custom TLS settings.
- Use `WithTransport` to set the HTTP transport.

`WithUserAgent` sets the `User-Agent` header. The public client has no option for other headers.

## Security

`WithAllowPrivateNetworks(false)` is the default. URL validation rejects localhost and private IP addresses before the initial request.

The default client does not validate redirect destinations or restrict connections to the IP addresses from that validation.
Do not treat this option as complete SSRF protection.

If you need to parse internal URLs in a trusted environment, set `WithAllowPrivateNetworks(true)`.

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

Allow private networks for a trusted internal tool:

```go
client := hermes.New(
    hermes.WithAllowPrivateNetworks(true),
)
```
