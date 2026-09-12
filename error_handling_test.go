package hermes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestErrorCodeClassification tests all error code paths.
func TestErrorCodeClassification(t *testing.T) {
	client := New(WithAllowPrivateNetworks(true)) // Allow localhost for testing network errors

	tests := []struct {
		name         string
		setupFunc    func() (string, context.Context)
		expectedCode ErrorCode
		blockPrivate bool
	}{
		{
			name: "ErrInvalidURL - empty URL",
			setupFunc: func() (string, context.Context) {
				return "", context.Background()
			},
			expectedCode: ErrInvalidURL,
		},
		{
			name: "ErrTimeout - context deadline exceeded",
			setupFunc: func() (string, context.Context) {
				// Create a server that delays longer than the context timeout
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(200 * time.Millisecond)
					w.WriteHeader(200)
					_, _ = w.Write([]byte("<html><body>Test</body></html>"))
				}))
				t.Cleanup(server.Close)

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				t.Cleanup(cancel)
				return server.URL, ctx
			},
			expectedCode: ErrTimeout,
		},
		{
			name: "ErrTimeout - context canceled",
			setupFunc: func() (string, context.Context) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
					w.WriteHeader(200)
					_, _ = w.Write([]byte("<html><body>Test</body></html>"))
				}))
				t.Cleanup(server.Close)

				ctx, cancel := context.WithCancel(context.Background())
				// Cancel immediately to trigger cancellation error
				cancel()
				return server.URL, ctx
			},
			expectedCode: ErrTimeout,
		},
		{
			name: "ErrSSRF - private network blocked",
			setupFunc: func() (string, context.Context) {
				return "http://192.168.1.1/test", context.Background()
			},
			expectedCode: ErrSSRF,
			blockPrivate: true,
		},
		{
			name: "ErrFetch - network error",
			setupFunc: func() (string, context.Context) {
				// Use a non-existent domain to trigger network error
				return "http://thisdoesnotexist.invalid/test", context.Background()
			},
			expectedCode: ErrFetch,
		},
		{
			name: "ErrFetch - connection refused",
			setupFunc: func() (string, context.Context) {
				// Use a local port that's not listening
				return "http://localhost:99999/test", context.Background()
			},
			expectedCode: ErrFetch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, ctx := tt.setupFunc()

			testClient := client
			if tt.blockPrivate {
				testClient = New()
			}

			result, err := testClient.Parse(ctx, url)
			assertParseFailure(t, result, err, tt.expectedCode, "Parse")
		})
	}
}

func assertParseFailure(t *testing.T, result *Result, err error, expectedCode ErrorCode, expectedOp string) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error, got none")
	}
	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("Expected ParseError, got: %T", err)
	}
	if parseErr.Code != expectedCode {
		t.Errorf("Expected error code %d (%s), got %d (%s). Error: %v",
			expectedCode, expectedCode.String(),
			parseErr.Code, parseErr.Code.String(), parseErr)
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got: %v", result)
	}
	if parseErr.Op != expectedOp {
		t.Errorf("Expected operation %q, got %q", expectedOp, parseErr.Op)
	}
}

// TestParseErrorMethods tests the ParseError interface methods.
func TestParseErrorMethods(t *testing.T) {
	originalErr := errors.New("original error")
	parseErr := &ParseError{
		Code: ErrFetch,
		URL:  "https://example.com",
		Op:   "Parse",
		Err:  originalErr,
	}

	// Test Error() method
	errStr := parseErr.Error()
	if !strings.Contains(errStr, "Parse") {
		t.Errorf("Error message should contain operation: %s", errStr)
	}
	if !strings.Contains(errStr, "https://example.com") {
		t.Errorf("Error message should contain URL: %s", errStr)
	}
	if !strings.Contains(errStr, "original error") {
		t.Errorf("Error message should contain original error: %s", errStr)
	}

	// Test Unwrap() method
	unwrapped := parseErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap should return original error, got: %v", unwrapped)
	}

	// Test Is() method
	anotherParseErr := &ParseError{Code: ErrFetch}
	if !parseErr.Is(anotherParseErr) {
		t.Error("Is() should return true for same error code")
	}

	differentCodeErr := &ParseError{Code: ErrTimeout}
	if parseErr.Is(differentCodeErr) {
		t.Error("Is() should return false for different error code")
	}

	if parseErr.Is(originalErr) {
		t.Error("Is() should return false for different error type")
	}
}

// TestErrorWrappingAndUnwrapping tests error wrapping behavior.
func TestErrorWrappingAndUnwrapping(t *testing.T) {
	client := New()

	// Test with invalid URL to get a ParseError
	result, err := client.Parse(context.Background(), "")

	if err == nil {
		t.Fatal("Expected error for empty URL")
	}

	// Check that it's a ParseError
	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("Expected ParseError, got: %T", err)
	}

	// Test errors.Is()
	if !errors.Is(err, &ParseError{Code: ErrInvalidURL}) {
		t.Error("errors.Is should work with ParseError")
	}

	// Test that result is nil on error
	if result != nil {
		t.Error("Result should be nil when there's an error")
	}
}

// TestParseHTMLErrorHandling tests error handling in ParseHTML.
func TestParseHTMLErrorHandling(t *testing.T) {
	client := New()
	ctx := context.Background()

	tests := []struct {
		name         string
		html         string
		url          string
		expectedCode ErrorCode
	}{
		{
			name:         "empty URL",
			html:         "<html><body>Test</body></html>",
			url:          "",
			expectedCode: ErrInvalidURL,
		},
		{
			name:         "empty HTML",
			html:         "",
			url:          "https://example.com",
			expectedCode: ErrInvalidURL,
		},
		{
			name:         "invalid URL format",
			html:         "<html><body>Test</body></html>",
			url:          "not-a-url",
			expectedCode: ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.ParseHTML(ctx, tt.html, tt.url)

			assertParseFailure(t, result, err, tt.expectedCode, "ParseHTML")
		})
	}
}

// TestContextCancellationErrorClassification tests that context cancellation is properly classified.
func TestContextCancellationErrorClassification(t *testing.T) {
	// Create a server that responds slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
		_, _ = w.Write([]byte("<html><body>Test content</body></html>"))
	}))
	defer server.Close()

	client := New(WithAllowPrivateNetworks(true)) // Allow localhost for testing

	t.Run("deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		result, err := client.Parse(ctx, server.URL)

		assertParseFailure(t, result, err, ErrTimeout, "Parse")
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel the context immediately
		cancel()

		result, err := client.Parse(ctx, server.URL)

		assertParseFailure(t, result, err, ErrTimeout, "Parse")
	})
}

// TestNetworkErrorClassification tests classification of various network errors.
func TestNetworkErrorClassification(t *testing.T) {
	client := New(WithAllowPrivateNetworks(true)) // Allow localhost for testing
	ctx := context.Background()

	tests := []struct {
		name         string
		url          string
		expectedCode ErrorCode
	}{
		{
			name:         "DNS resolution failure",
			url:          "http://definitely-does-not-exist.invalid",
			expectedCode: ErrFetch,
		},
		{
			name:         "Connection refused",
			url:          "http://localhost:99999",
			expectedCode: ErrFetch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.Parse(ctx, tt.url)

			assertParseFailure(t, result, err, tt.expectedCode, "Parse")
		})
	}
}

// TestSSRFProtectionNetworkErrors tests that SSRF protection properly blocks private networks.
func TestSSRFProtectionNetworkErrors(t *testing.T) {
	// Use default client (no private networks allowed)
	client := New()
	ctx := context.Background()

	tests := []struct {
		name         string
		url          string
		expectedCode ErrorCode
	}{
		{
			name:         "Private network blocked",
			url:          "http://192.168.1.1",
			expectedCode: ErrSSRF,
		},
		{
			name:         "Localhost blocked",
			url:          "http://127.0.0.1:8080",
			expectedCode: ErrSSRF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.Parse(ctx, tt.url)

			assertParseFailure(t, result, err, tt.expectedCode, "Parse")
		})
	}
}

// TestErrorCodeValues tests that error codes have expected values.
func TestErrorCodeValues(t *testing.T) {
	expectedCodes := map[ErrorCode]string{
		ErrInvalidURL: "invalid URL",
		ErrFetch:      "fetch error",
		ErrTimeout:    "timeout",
		ErrSSRF:       "SSRF blocked",
		ErrExtract:    "extraction error",
		ErrContext:    "context cancelled",
	}

	for code, expectedStr := range expectedCodes {
		codeStr := code.String()
		if codeStr == "" {
			t.Errorf("Error code %d should have a string representation", code)
		}
		if codeStr != expectedStr {
			t.Errorf("Error code %d should return '%s', got: '%s'", code, expectedStr, codeStr)
		}
	}
}

// BenchmarkErrorClassification benchmarks the error classification performance.
func BenchmarkErrorClassification(b *testing.B) {
	client := New()
	ctx := context.Background()

	b.Run("InvalidURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := client.Parse(ctx, "")
			if err == nil {
				b.Fatal("Expected error")
			}
		}
	})

	b.Run("TimeoutError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
			_, err := client.Parse(ctx, "http://example.com")
			cancel()
			if err == nil {
				b.Fatal("Expected timeout error")
			}
		}
	})
}
