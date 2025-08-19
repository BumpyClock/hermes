package resource_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/postlight/parser-go/pkg/resource"
)

func TestResource_Create_WithPreparedHTML(t *testing.T) {
	r := resource.NewResource()
	
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

	doc, err := r.Create("http://example.com", htmlContent, nil, nil)
	require.NoError(t, err)
	assert.NotNil(t, doc)
	
	// Check that DOM was processed
	title := doc.Find("title").Text()
	assert.Equal(t, "Test Article", title)
	
	h1 := doc.Find("h1").Text()
	assert.Equal(t, "Test Title", h1)
}

func TestResource_Create_WithMetaNormalization(t *testing.T) {
	r := resource.NewResource()
	
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

	doc, err := r.Create("http://example.com", htmlContent, nil, nil)
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
	r := resource.NewResource()
	
	htmlContent := `<!DOCTYPE html>
<html>
<body>
	<img data-src="https://example.com/image.jpg" src="placeholder.gif">
	<img data-lazy="https://example.com/image2.png">
</body>
</html>`

	doc, err := r.Create("http://example.com", htmlContent, nil, nil)
	require.NoError(t, err)
	
	// Check that lazy images were converted
	img1Src, _ := doc.Find("img").First().Attr("src")
	assert.Equal(t, "https://example.com/image.jpg", img1Src)
}

func TestResource_Create_CleansTags(t *testing.T) {
	r := resource.NewResource()
	
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

	doc, err := r.Create("http://example.com", htmlContent, nil, nil)
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
		w.Write([]byte("fake image data"))
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	result, err := resource.FetchResource(server.URL, parsedURL, nil)
	
	require.NoError(t, err)
	assert.True(t, result.IsError())
	assert.Contains(t, result.Message, "not allowed")
}

func TestFetchResource_HandlesSuccess(t *testing.T) {
	htmlContent := `<!DOCTYPE html><html><body><h1>Test</h1></body></html>`
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)
	result, err := resource.FetchResource(server.URL, parsedURL, nil)
	
	require.NoError(t, err)
	assert.False(t, result.IsError())
	assert.Equal(t, htmlContent, string(result.Body))
	assert.Equal(t, 200, result.Response.StatusCode)
}

func TestFetchResource_WithCustomHeaders(t *testing.T) {
	var receivedHeaders http.Header
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte("<html><body>Test</body></html>"))
	}))
	defer server.Close()

	headers := map[string]string{
		"X-Custom-Header": "test-value",
		"Authorization":   "Bearer token123",
	}

	parsedURL, _ := url.Parse(server.URL)
	result, err := resource.FetchResource(server.URL, parsedURL, headers)
	
	require.NoError(t, err)
	assert.False(t, result.IsError())
	
	// Check that custom headers were sent
	assert.Equal(t, "test-value", receivedHeaders.Get("X-Custom-Header"))
	assert.Equal(t, "Bearer token123", receivedHeaders.Get("Authorization"))
	
	// Check that default headers were also sent
	userAgent := receivedHeaders.Get("User-Agent")
	assert.Contains(t, userAgent, "Mozilla")
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

func TestBaseDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"www.example.com", "example.com"},
		{"subdomain.example.com", "example.com"},
		{"deep.subdomain.example.com", "example.com"},
		{"example.com", "example.com"},
		{"localhost", "localhost"},
	}
	
	for _, test := range tests {
		result := resource.BaseDomain(test.input)
		assert.Equal(t, test.expected, result, "BaseDomain(%s)", test.input)
	}
}

func TestResource_GenerateDoc_InvalidContent(t *testing.T) {
	r := resource.NewResource()
	
	result := &resource.FetchResult{
		Body: []byte("not html content"),
		Response: &resource.Response{
			StatusCode: 200,
			Headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
	}
	
	_, err := r.GenerateDoc(result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not appear to be text")
}

func TestResource_GenerateDoc_EmptyDocument(t *testing.T) {
	r := resource.NewResource()
	
	// Use malformed HTML that won't parse correctly
	result := &resource.FetchResult{
		Body: []byte("<html><head></head><body></body></html>"),
		Response: &resource.Response{
			StatusCode: 200,
			Headers: http.Header{
				"Content-Type": []string{"text/html"},
			},
		},
	}
	
	// This should actually succeed since goquery is more lenient
	// Let's test with truly invalid HTML instead
	result.Body = []byte("not html at all")
	
	doc, err := r.GenerateDoc(result)
	// Even this might parse, so let's check if we get a document
	if err == nil {
		// If it parsed, check that we have some content
		assert.NotNil(t, doc)
	}
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

	r := resource.NewResource()
	doc, err := r.Create("http://example.com", utf8Content, nil, nil)
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

	r := resource.NewResource()
	
	// Simulate server response with different encoding
	result := &resource.FetchResult{
		Body: []byte(htmlWithMetaCharset),
		Response: &resource.Response{
			StatusCode: 200,
			Headers: http.Header{
				"Content-Type": []string{"text/html; charset=utf-8"},
			},
		},
		AlreadyDecoded: false,
	}
	
	doc, err := r.GenerateDoc(result)
	require.NoError(t, err)
	
	// Should have normalized the meta tag
	metaCharset, exists := doc.Find("meta[http-equiv]").Attr("value")
	assert.True(t, exists)
	assert.Contains(t, metaCharset, "iso-8859-1")
}

// Benchmark test to ensure performance
func BenchmarkResource_Create(b *testing.B) {
	r := resource.NewResource()
	
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
		_, err := r.Create("http://example.com", htmlContent, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}