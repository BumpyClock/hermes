package resource_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BumpyClock/hermes/internal/resource"
)

func TestHTTPClientGet(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check default headers (using JS-compatible User-Agent)
		assert.Contains(t, r.Header.Get("User-Agent"), "Mozilla")
		assert.Equal(t, "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", r.Header.Get("Accept"))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>Test content</body></html>"))
	}))
	defer server.Close()

	client := resource.CreateDefaultHTTPClient()
	resp, err := client.Get(context.Background(), server.URL)

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
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	headers := map[string]string{
		"X-Custom": customHeader,
	}

	client := resource.CreateDefaultHTTPClient()
	client.Headers = headers
	resp, err := client.Get(context.Background(), server.URL)

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
		_, _ = w.Write([]byte("Success"))
	}))
	defer server.Close()

	client := resource.CreateDefaultHTTPClient()
	resp, err := client.GetWithRetry(context.Background(), server.URL, 3)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 3, attemptCount)
}

func TestHTTPClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	client := resource.CreateDefaultHTTPClient()
	_, err := client.Get(ctx, server.URL)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestHTTPClientError(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := resource.CreateDefaultHTTPClient()
	_, err := client.Get(context.Background(), server.URL)

	require.EqualError(t, err, "failed after 4 attempts: HTTP 404: 404 Not Found")
	assert.Equal(t, 1, attemptCount)
}

func TestHTTPClientResponseBodyReadFailure(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		wantError  string
		wrappedEOF bool
	}{
		{"success", http.StatusOK, "failed after 1 attempts: reading response body: unexpected EOF", true},
		{"client error", http.StatusNotFound, "failed after 1 attempts: HTTP 404: 404 Not Found (failed to read error response)", false},
		{"server error", http.StatusInternalServerError, "failed after 1 attempts: HTTP 500: 500 Internal Server Error (failed to read error response)", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "100")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte("truncated"))
			}))
			defer server.Close()

			client := resource.CreateDefaultHTTPClient()
			resp, err := client.GetWithRetry(context.Background(), server.URL, 0)

			require.EqualError(t, err, tc.wantError)
			assert.Nil(t, resp)
			assert.Equal(t, tc.wrappedEOF, errors.Is(err, io.ErrUnexpectedEOF))
		})
	}
}
