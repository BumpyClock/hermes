// Package hermes provides a high-performance web content extraction library
// that transforms web pages into clean, structured data.
//
// Hermes extracts article content, titles, authors, dates, images, and more
// from any URL using site-specific custom parsers and generic fallback extraction.
//
// # Basic Usage
//
// Create a client and parse a URL:
//
//	client := hermes.New()
//	result, err := client.Parse(context.Background(), "https://example.com/article")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Title)
//	fmt.Println(result.Content)
//
// # Configuration
//
// The client can be configured with various options:
//
//	client := hermes.New(
//	    hermes.WithTimeout(30 * time.Second),
//	    hermes.WithUserAgent("MyApp/1.0"),
//	    hermes.WithAllowPrivateNetworks(false),
//	)
//
// # Custom HTTP Client
//
// You can provide your own HTTP client for custom transport settings:
//
//	httpClient := &http.Client{
//	    Transport: &http.Transport{
//	        Proxy: http.ProxyFromEnvironment,
//	        MaxIdleConns: 100,
//	    },
//	}
//	client := hermes.New(hermes.WithHTTPClient(httpClient))
//
// # Parsing Pre-fetched HTML
//
// If you already have the HTML content, you can parse it directly:
//
//	html := "<html>...</html>"
//	result, err := client.ParseHTML(context.Background(), html, "https://example.com")
//
// # Error Handling
//
// Errors are typed for programmatic handling:
//
//	result, err := client.Parse(ctx, url)
//	if err != nil {
//	    var parseErr *hermes.ParseError
//	    if errors.As(err, &parseErr) {
//	        switch parseErr.Code {
//	        case hermes.ErrFetch:
//	            // Handle fetch error
//	        case hermes.ErrTimeout:
//	            // Handle timeout
//	        case hermes.ErrSSRF:
//	            // Handle SSRF protection
//	        }
//	    }
//	}
//
// # Thread Safety
//
// The Client is thread-safe and should be reused across goroutines.
// Create one client and share it throughout your application.
//
// # Concurrency
//
// The library parses one URL at a time. For concurrent parsing,
// implement your own worker pool:
//
//	var wg sync.WaitGroup
//	sem := make(chan struct{}, 10) // Limit concurrency
//	
//	for _, url := range urls {
//	    wg.Add(1)
//	    sem <- struct{}{}
//	    
//	    go func(u string) {
//	        defer wg.Done()
//	        defer func() { <-sem }()
//	        
//	        result, err := client.Parse(ctx, u)
//	        // Handle result
//	    }(url)
//	}
//	wg.Wait()
package hermes