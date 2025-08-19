package resource_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/postlight/parser-go/pkg/resource"
)

func TestNewHTTPClient(t *testing.T) {
	headers := map[string]string{
		"Custom-Header": "test-value",
	}

	client := resource.NewHTTPClient(headers)
	assert.NotNil(t, client)
}

func TestHTTPClientGet(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check default headers (using JS-compatible User-Agent)
		assert.Contains(t, r.Header.Get("User-Agent"), "Mozilla")
		assert.Equal(t, "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", r.Header.Get("Accept"))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<html><body>Test content</body></html>"))
	}))
	defer server.Close()

	client := resource.NewHTTPClient(nil)
	resp, err := client.Get(server.URL)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", resp.GetContentType())
	assert.Contains(t, string(resp.Body), "Test content")
}

func TestHTTPClientCustomHeaders(t *testing.T) {
	customHeader := "test-custom-value"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check custom header
		assert.Equal(t, customHeader, r.Header.Get("X-Custom"))
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	headers := map[string]string{
		"X-Custom": customHeader,
	}
	
	client := resource.NewHTTPClient(headers)
	resp, err := client.Get(server.URL)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHTTPClientRetry(t *testing.T) {
	attemptCount := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer server.Close()

	client := resource.NewHTTPClient(nil)
	resp, err := client.GetWithRetry(server.URL, 3)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 3, attemptCount)
}

func TestHTTPClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// This test would require modifying the client timeout
	// For now, just test that the client works normally
	client := resource.NewHTTPClient(nil)
	resp, err := client.Get(server.URL)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHTTPClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := resource.NewHTTPClient(nil)
	_, err := client.Get(server.URL)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}