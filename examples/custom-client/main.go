// Package main demonstrates advanced HTTP client configuration with Hermes.
//
// This example shows how to:
// - Create custom HTTP clients with specific configurations
// - Configure connection pooling and timeouts
// - Set up proxy support
// - Configure TLS settings
// - Use custom transport options
// - Handle different authentication scenarios
//
// Run with: go run examples/custom-client/main.go
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/BumpyClock/hermes"
)

func main() {
	fmt.Println("Hermes Custom HTTP Client Example")
	fmt.Println("=================================")

	testURL := "https://httpbin.org/headers"

	// Example 1: Basic custom client with connection pooling
	fmt.Println("\n1. Custom Client with Connection Pooling")
	fmt.Println("----------------------------------------")
	basicCustomClient(testURL)

	// Example 2: Client with proxy support
	fmt.Println("\n2. Client with Proxy Support")
	fmt.Println("----------------------------")
	proxyClient(testURL)

	// Example 3: Client with custom TLS settings
	fmt.Println("\n3. Client with Custom TLS Settings")
	fmt.Println("----------------------------------")
	tlsClient(testURL)

	// Example 4: High-performance client for batch processing
	fmt.Println("\n4. High-Performance Batch Client")
	fmt.Println("--------------------------------")
	highPerformanceClient(testURL)

	// Example 5: Client with custom headers and authentication
	fmt.Println("\n5. Client with Custom Headers")
	fmt.Println("-----------------------------")
	customHeadersClient(testURL)

	fmt.Println("\n🎉 All examples completed!")
}

// basicCustomClient demonstrates basic HTTP client customization
func basicCustomClient(testURL string) {
	// Create custom HTTP client with specific settings
	httpClient := &http.Client{
		Timeout: 45 * time.Second,
		Transport: &http.Transport{
			// Connection settings
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			
			// Dial settings
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			
			// Response settings
			ResponseHeaderTimeout: 10 * time.Second,
			DisableCompression:    false,
		},
	}

	// Create Hermes client with custom HTTP client
	client := hermes.New(
		hermes.WithHTTPClient(httpClient),
		hermes.WithUserAgent("CustomClient/1.0"),
		hermes.WithContentType("text"),
	)

	parseAndDisplay(client, testURL, "Basic Custom Client")
}

// proxyClient demonstrates HTTP client with proxy support
func proxyClient(testURL string) {
	// Note: This example shows proxy configuration but doesn't use a real proxy
	// Uncomment and modify the proxy URL if you have a proxy server
	
	// proxyURL, _ := url.Parse("http://proxy.example.com:8080")
	
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			// Proxy configuration (commented out for demo)
			// Proxy: http.ProxyURL(proxyURL),
			
			// For demo, we'll use ProxyFromEnvironment which checks env vars
			Proxy: func(req *http.Request) (*url.URL, error) {
				// In real usage, return proxyURL for proxy routing
				// For this demo, we'll not use a proxy
				fmt.Printf("🔄 Proxy check for: %s (no proxy configured)\n", req.URL.Host)
				return nil, nil // No proxy
			},
			
			MaxIdleConns:    20,
			IdleConnTimeout: 60 * time.Second,
		},
	}

	client := hermes.New(
		hermes.WithHTTPClient(httpClient),
		hermes.WithUserAgent("ProxyClient/1.0"),
	)

	parseAndDisplay(client, testURL, "Proxy Client (demo)")
}

// tlsClient demonstrates custom TLS configuration
func tlsClient(testURL string) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// Security settings
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
				
				// Certificate verification (be careful with these in production)
				InsecureSkipVerify: false,
				ServerName:         "", // Leave empty to use hostname from URL
				
				// Cipher suite preferences (optional)
				PreferServerCipherSuites: true,
			},
			
			// TLS handshake timeout
			TLSHandshakeTimeout: 10 * time.Second,
			
			MaxIdleConns:    30,
			IdleConnTimeout: 90 * time.Second,
		},
	}

	client := hermes.New(
		hermes.WithHTTPClient(httpClient),
		hermes.WithUserAgent("TLSClient/1.0"),
		hermes.WithContentType("html"),
	)

	parseAndDisplay(client, testURL, "Custom TLS Client")
}

// highPerformanceClient demonstrates optimized client for batch processing
func highPerformanceClient(testURL string) {
	httpClient := &http.Client{
		Timeout: 20 * time.Second, // Shorter timeout for batch processing
		Transport: &http.Transport{
			// Optimized for high throughput
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     120 * time.Second,
			
			// Faster connection establishment
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			
			// Optimized timeouts
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			
			// Connection reuse
			DisableKeepAlives: false,
			ForceAttemptHTTP2: true,
		},
	}

	client := hermes.New(
		hermes.WithHTTPClient(httpClient),
		hermes.WithUserAgent("HighPerformanceBot/2.0"),
		hermes.WithContentType("markdown"),
	)

	fmt.Printf("🚀 High-performance client configured for batch processing\n")
	parseAndDisplay(client, testURL, "High-Performance Client")
}

// customHeadersClient demonstrates client with custom transport for headers
func customHeadersClient(testURL string) {
	// Create a custom transport that adds authentication headers
	baseTransport := &http.Transport{
		MaxIdleConns:    30,
		IdleConnTimeout: 60 * time.Second,
	}

	// Wrap transport to add custom headers
	customTransport := &customHeaderTransport{
		Transport: baseTransport,
		Headers: map[string]string{
			"X-API-Key":        "demo-api-key-12345",
			"X-Client-Version": "1.0.0",
			"Accept":           "text/html,application/xhtml+xml",
		},
	}

	httpClient := &http.Client{
		Timeout:   25 * time.Second,
		Transport: customTransport,
	}

	client := hermes.New(
		hermes.WithHTTPClient(httpClient),
		hermes.WithUserAgent("AuthenticatedClient/1.0"),
	)

	parseAndDisplay(client, testURL, "Custom Headers Client")
}

// customHeaderTransport wraps an http.Transport to add custom headers
type customHeaderTransport struct {
	Transport http.RoundTripper
	Headers   map[string]string
}

func (t *customHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add custom headers to the request
	for key, value := range t.Headers {
		req.Header.Set(key, value)
	}
	
	fmt.Printf("🔧 Added %d custom headers to request\n", len(t.Headers))
	
	// Use the wrapped transport
	return t.Transport.RoundTrip(req)
}

// parseAndDisplay parses a URL and displays the results
func parseAndDisplay(client *hermes.Client, testURL, clientName string) {
	fmt.Printf("Testing with %s...\n", clientName)
	
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	result, err := client.Parse(ctx, testURL)
	duration := time.Since(start)

	if err != nil {
		if parseErr, ok := err.(*hermes.ParseError); ok {
			fmt.Printf("❌ Error [%s]: %v (took %v)\n", parseErr.Code, parseErr.Err, duration)
		} else {
			fmt.Printf("❌ Error: %v (took %v)\n", err, duration)
		}
		return
	}

	fmt.Printf("✅ Success! (took %v)\n", duration)
	fmt.Printf("   Title: %s\n", truncate(result.Title, 50))
	fmt.Printf("   Domain: %s\n", result.Domain)
	fmt.Printf("   Word Count: %d\n", result.WordCount)
	if result.Content != "" {
		fmt.Printf("   Content Preview: %s\n", truncate(result.Content, 80))
	}
}

// truncate shortens a string to the specified length with ellipsis
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}