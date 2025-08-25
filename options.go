package hermes

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the Client
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client for the parser.
// This allows you to configure connection pooling, timeouts, proxies, etc.
//
// Example:
//
//	httpClient := &http.Client{
//	    Timeout: 60 * time.Second,
//	    Transport: &http.Transport{
//	        MaxIdleConns: 200,
//	    },
//	}
//	client := hermes.New(hermes.WithHTTPClient(httpClient))
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTransport sets a custom HTTP transport for the parser.
// This is useful for configuring proxies, TLS settings, connection pooling, etc.
// If both WithHTTPClient and WithTransport are used, WithHTTPClient takes precedence.
//
// Example:
//
//	transport := &http.Transport{
//	    Proxy: http.ProxyFromEnvironment,
//	    MaxIdleConns: 100,
//	    IdleConnTimeout: 90 * time.Second,
//	}
//	client := hermes.New(hermes.WithTransport(transport))
func WithTransport(transport http.RoundTripper) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Transport = transport
	}
}

// WithTimeout sets the timeout for HTTP requests.
// This timeout applies to the entire request, including connection time,
// redirects, and reading the response body.
//
// Example:
//
//	client := hermes.New(hermes.WithTimeout(30 * time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
	}
}

// WithUserAgent sets the User-Agent header for HTTP requests.
// This is useful for identifying your application to web servers.
//
// Example:
//
//	client := hermes.New(hermes.WithUserAgent("MyApp/1.0"))
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithAllowPrivateNetworks allows or disallows parsing of private network URLs.
// By default, private networks are blocked for security (SSRF protection).
// Set to true only in trusted environments where you need to parse internal URLs.
//
// Private networks include:
//   - 10.0.0.0/8
//   - 172.16.0.0/12
//   - 192.168.0.0/16
//   - 127.0.0.0/8 (localhost)
//   - ::1 (IPv6 localhost)
//   - fc00::/7 (IPv6 private)
//
// Example:
//
//	// For internal tools that need to parse intranet content
//	client := hermes.New(hermes.WithAllowPrivateNetworks(true))
func WithAllowPrivateNetworks(allow bool) Option {
	return func(c *Client) {
		c.allowPrivateNetworks = allow
	}
}

// WithContentType sets the output content type for parsing.
// Valid options are "html", "markdown", and "text".
// By default, content is returned as HTML.
//
// Example:
//
//	// Get content as markdown
//	client := hermes.New(hermes.WithContentType("markdown"))
func WithContentType(contentType string) Option {
	return func(c *Client) {
		c.contentType = contentType
	}
}