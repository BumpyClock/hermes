package resource_test

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/internal/resource"
)

func TestResource_Create_WithPreparedHTML(t *testing.T) {

	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Test Article</title>
</head>
<body>
	<h1>Test Title</h1>
	<p>Test content</p>
</body>
</html>`

	doc, err := resource.CreateDocument(context.Background(), "http://example.com", htmlContent, nil, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, doc)

	// Check that DOM was processed
	title := doc.Find("title").Text()
	assert.Equal(t, "Test Article", title)

	h1 := doc.Find("h1").Text()
	assert.Equal(t, "Test Title", h1)
}

func TestResource_Create_WithMetaNormalization(t *testing.T) {

	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta property="og:title" content="OpenGraph Title">
	<meta name="description" content="Meta Description">
</head>
<body>
	<p>Content</p>
</body>
</html>`

	doc, err := resource.CreateDocument(context.Background(), "http://example.com", htmlContent, nil, nil, nil)
	require.NoError(t, err)

	// Check that property was converted to name
	ogTitle, exists := doc.Find("meta[name='og:title']").Attr("value")
	assert.True(t, exists)
	assert.Equal(t, "OpenGraph Title", ogTitle)

	// Check that content was converted to value
	description, exists := doc.Find("meta[name='description']").Attr("value")
	assert.True(t, exists)
	assert.Equal(t, "Meta Description", description)
}

func TestResource_Create_WithLazyImages(t *testing.T) {

	htmlContent := `<!DOCTYPE html>
<html>
<body>
	<img data-src="https://example.com/image.jpg" src="placeholder.gif">
	<img data-lazy="https://example.com/image2.png">
</body>
</html>`

	doc, err := resource.CreateDocument(context.Background(), "http://example.com", htmlContent, nil, nil, nil)
	require.NoError(t, err)

	// Check that lazy images were converted
	img1Src, _ := doc.Find("img").First().Attr("src")
	assert.Equal(t, "https://example.com/image.jpg", img1Src)
}

// Mercury's convertLazyLoadedImages visits cheerio attributes in source order,
// so the last matching attribute sets src or srcset.
func TestConvertLazyLoadedImagesLastMatchInSourceOrderWins(t *testing.T) {
	tests := []struct {
		name       string
		html       string
		wantSrc    string
		wantSrcset string
	}{
		{
			name: "src candidates",
			html: `<img src="placeholder.gif" data-hi-res-src="https://example.com/hi.jpg" ` +
				`data-low-res-src="https://example.com/low.jpg" data-raw-src="https://example.com/raw.jpg">`,
			wantSrc: "https://example.com/raw.jpg",
		},
		{
			name: "srcset candidates",
			html: `<img data-srcset="https://example.com/a.jpg 1x" data-lazy-srcset="https://example.com/b.jpg 2x" ` +
				`data-hidpi-srcset="https://example.com/c.jpg 480w">`,
			wantSrcset: "https://example.com/c.jpg 480w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Repeat because the old map-driven choice changed between runs.
			for range 25 {
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
				require.NoError(t, err)

				resource.ConvertLazyLoadedImages(doc)

				img := doc.Find("img")
				src, _ := img.Attr("src")
				srcset, _ := img.Attr("srcset")
				require.Equal(t, tt.wantSrc, src)
				require.Equal(t, tt.wantSrcset, srcset)
			}
		})
	}
}

func TestResource_Create_CleansTags(t *testing.T) {

	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<script>alert('test');</script>
	<style>body { color: red; }</style>
</head>
<body>
	<p>Content</p>
	<form><input type="text"></form>
	<!-- This is a comment -->
</body>
</html>`

	doc, err := resource.CreateDocument(context.Background(), "http://example.com", htmlContent, nil, nil, nil)
	require.NoError(t, err)

	// Check that unwanted tags were removed
	assert.Equal(t, 0, doc.Find("script").Length())
	assert.Equal(t, 0, doc.Find("style").Length())
	assert.Equal(t, 0, doc.Find("form").Length())
}

func TestFetchResource_ValidatesResponse(t *testing.T) {
	// Test server returning bad content type
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("fake image data"))
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	result, err := resource.Fetch(context.Background(), server.URL, parsedURL, nil, resource.CreateDefaultHTTPClient())

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestFetchResource_HandlesSuccess(t *testing.T) {
	htmlContent := `<!DOCTYPE html><html><body><h1>Test</h1></body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	result, err := resource.Fetch(context.Background(), server.URL, parsedURL, nil, resource.CreateDefaultHTTPClient())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, htmlContent, string(result.Body))
	assert.Equal(t, 200, result.StatusCode)
}

func TestFetchResource_WithCustomHeaders(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("<html><body>Test</body></html>"))
	}))
	defer server.Close()

	headers := map[string]string{
		"X-Custom-Header": "test-value",
		"Authorization":   "Bearer token123",
	}

	parsedURL, _ := url.Parse(server.URL)
	result, err := resource.Fetch(context.Background(), server.URL, parsedURL, headers, resource.CreateDefaultHTTPClient())

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Check that custom headers were sent
	assert.Equal(t, "test-value", receivedHeaders.Get("X-Custom-Header"))
	assert.Equal(t, "Bearer token123", receivedHeaders.Get("Authorization"))

	// Check that default headers were also sent
	userAgent := receivedHeaders.Get("User-Agent")
	assert.Contains(t, userAgent, "Mozilla")
}

func TestFetchResource_PreservesClientHeaders(t *testing.T) {
	defaultClientHeaders := map[string]string{
		"X-Client-Only": "preserved",
		"X-Shared":      "client",
		"User-Agent":    "ClientAgent",
	}
	for _, test := range []struct {
		name          string
		clientHeaders map[string]string
		headers       map[string]string
		wantClient    string
		wantValue     string
		wantAgent     string
	}{
		{"client only", defaultClientHeaders, nil, "preserved", "client", "ClientAgent"},
		{"request override", defaultClientHeaders, map[string]string{"X-Shared": "request", "User-Agent": "RequestAgent"}, "preserved", "request", "RequestAgent"},
		{"case-insensitive override", defaultClientHeaders, map[string]string{"x-shared": "request", "user-agent": "RequestAgent"}, "preserved", "request", "RequestAgent"},
		{"lowercase client headers", map[string]string{"x-client-only": "preserved", "x-shared": "client", "user-agent": "ClientAgent"}, nil, "preserved", "client", "ClientAgent"},
		{"nil client headers", nil, map[string]string{"X-Shared": "request", "User-Agent": "RequestAgent"}, "", "request", "RequestAgent"},
	} {
		t.Run(test.name, func(t *testing.T) {
			received := make(chan http.Header, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received <- r.Header.Clone()
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte("<html><body>Test</body></html>"))
			}))
			defer server.Close()

			client := resource.CreateDefaultHTTPClient()
			client.Headers = test.clientHeaders
			defer client.Client.CloseIdleConnections()
			clientHeaders := maps.Clone(client.Headers)
			requestHeaders := maps.Clone(test.headers)

			result, err := resource.Fetch(context.Background(), server.URL, nil, test.headers, client)
			require.NoError(t, err)
			require.NotNil(t, result)

			headers := <-received
			assert.Equal(t, test.wantClient, headers.Get("X-Client-Only"))
			assert.Equal(t, test.wantValue, headers.Get("X-Shared"))
			assert.Equal(t, test.wantAgent, headers.Get("User-Agent"))
			assert.Equal(t, "en-US,en;q=0.5", headers.Get("Accept-Language"))
			assert.Equal(t, clientHeaders, client.Headers)
			assert.Equal(t, requestHeaders, test.headers)
		})
	}
}

func TestFetchResource_RequestHeadersDoNotLeakToNextRequest(t *testing.T) {
	received := make(chan http.Header, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Clone()
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>Test</body></html>"))
	}))
	defer server.Close()

	client := resource.CreateDefaultHTTPClient()
	client.Headers = map[string]string{
		"X-Shared":   "client",
		"User-Agent": "ClientAgent",
	}
	defer client.Client.CloseIdleConnections()
	clientHeaders := maps.Clone(client.Headers)

	override := map[string]string{"x-shared": "request", "user-agent": "RequestAgent"}
	_, err := resource.Fetch(context.Background(), server.URL, nil, override, client)
	require.NoError(t, err)

	_, err = resource.Fetch(context.Background(), server.URL, nil, nil, client)
	require.NoError(t, err)

	first, second := <-received, <-received
	assert.Equal(t, "request", first.Get("X-Shared"))
	assert.Equal(t, "RequestAgent", first.Get("User-Agent"))
	assert.Equal(t, "client", second.Get("X-Shared"))
	assert.Equal(t, "ClientAgent", second.Get("User-Agent"))
	assert.Equal(t, clientHeaders, client.Headers)
	assert.Equal(t, map[string]string{"x-shared": "request", "user-agent": "RequestAgent"}, override)
}

func TestValidateResponse_ContentLength(t *testing.T) {
	response := &resource.Response{
		StatusCode: 200,
		Headers: http.Header{
			"Content-Type":   []string{"text/html"},
			"Content-Length": []string{"10485760"}, // 10MB > 5MB limit
		},
	}

	err := resource.ValidateResponse(response, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too large")
}

func TestValidateResponse_NonOKStatus(t *testing.T) {
	response := &resource.Response{
		StatusCode: 404,
		Headers: http.Header{
			"Content-Type": []string{"text/html"},
		},
	}

	// Should fail with parseNon200=false
	err := resource.ValidateResponse(response, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")

	// Should pass with parseNon200=true
	err = resource.ValidateResponse(response, true)
	assert.NoError(t, err)
}

func TestResource_GenerateDoc_InvalidContent(t *testing.T) {
	_, err := resource.PrepareDocument(context.Background(), []byte("not html content"), "application/json", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not appear to be text")
}

func TestPrepareDocument_BareText(t *testing.T) {
	doc, err := resource.PrepareDocument(context.Background(), []byte("not html at all"), "text/html", false)
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "not html at all", doc.Find("body").Text())
}

func TestEncodingDetection(t *testing.T) {
	// Test with UTF-8 content
	utf8Content := `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>UTF-8 Test</title>
</head>
<body>
	<p>Unicode: ñáéíóú</p>
</body>
</html>`

	doc, err := resource.CreateDocument(context.Background(), "http://example.com", utf8Content, nil, nil, nil)
	require.NoError(t, err)

	title := doc.Find("title").Text()
	assert.Equal(t, "UTF-8 Test", title)

	content := doc.Find("p").Text()
	assert.Contains(t, content, "ñáéíóú")
}

func TestResource_Create_EncodingMismatch(t *testing.T) {
	// HTML that declares ISO-8859-1 in meta tag but is served as UTF-8
	htmlWithMetaCharset := `<!DOCTYPE html>
<html>
<head>
	<meta http-equiv="content-type" content="text/html; charset=iso-8859-1">
	<title>Encoding Test</title>
</head>
<body>
	<p>Test content</p>
</body>
</html>`

	// Simulate server response with different encoding
	doc, err := resource.PrepareDocument(context.Background(), []byte(htmlWithMetaCharset), "text/html; charset=utf-8", false)
	require.NoError(t, err)

	// Should have normalized the meta tag
	metaCharset, exists := doc.Find("meta[http-equiv]").Attr("value")
	assert.True(t, exists)
	assert.Contains(t, metaCharset, "iso-8859-1")
}

// Benchmark test to ensure performance.
func BenchmarkResource_Create(b *testing.B) {

	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Benchmark Test</title>
</head>
<body>
	<h1>Title</h1>
	<p>Content paragraph with some text.</p>
	<div class="container">
		<p>More content</p>
		<img src="image.jpg" alt="test">
	</div>
</body>
</html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resource.CreateDocument(context.Background(), "http://example.com", htmlContent, nil, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestCreateDocument_FetchErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		url  string
		want string
	}{
		{name: "missing client", url: "https://example.com", want: "resource fetch failed: HTTP client is required"},
		{name: "invalid URL", url: ":", want: `resource fetch failed: Invalid URL: parse ":": missing protocol scheme`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := resource.CreateDocument(context.Background(), tc.url, "", nil, nil, nil)
			require.EqualError(t, err, tc.want)
			assert.Nil(t, doc)
			assert.Nil(t, errors.Unwrap(err))
		})
	}
}

func TestCreateDocument_PredecodedHTML(t *testing.T) {
	doc, err := resource.CreateDocument(context.Background(), "https://example.com", `<html><head><meta charset="windows-1252"></head><body><p>café 日本語</p></body></html>`, nil, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "café 日本語", doc.Find("p").Text())
}

func TestPrepareDocument_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	doc, err := resource.PrepareDocument(ctx, []byte("<p>content</p>"), "text/html", true)
	require.EqualError(t, err, "document processing timed out")
	assert.Nil(t, doc)
	assert.Nil(t, errors.Unwrap(err))
}
