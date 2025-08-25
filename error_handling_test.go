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

// TestErrorCodeClassification tests all error code paths
func TestErrorCodeClassification(t *testing.T) {
	client := New(WithAllowPrivateNetworks(true)) // Allow localhost for testing network errors

	tests := []struct {
		name         string
		setupFunc    func() (string, context.Context)
		expectedCode ErrorCode
		shouldError  bool
	}{
		{
			name: "ErrInvalidURL - empty URL",
			setupFunc: func() (string, context.Context) {
				return "", context.Background()
			},
			expectedCode: ErrInvalidURL,
			shouldError:  true,
		},
		{
			name: "ErrTimeout - context deadline exceeded",
			setupFunc: func() (string, context.Context) {
				// Create a server that delays longer than the context timeout
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(200 * time.Millisecond)
					w.WriteHeader(200)
					w.Write([]byte("<html><body>Test</body></html>"))
				}))
				t.Cleanup(server.Close)

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				t.Cleanup(cancel)
				return server.URL, ctx
			},
			expectedCode: ErrTimeout,
			shouldError:  true,
		},
		{
			name: "ErrTimeout - context canceled",
			setupFunc: func() (string, context.Context) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
					w.WriteHeader(200)
					w.Write([]byte("<html><body>Test</body></html>"))
				}))
				t.Cleanup(server.Close)

				ctx, cancel := context.WithCancel(context.Background())
				// Cancel immediately to trigger cancellation error
				cancel()
				return server.URL, ctx
			},
			expectedCode: ErrTimeout,
			shouldError:  true,
		},
		{
			name: "ErrSSRF - private network blocked",
			setupFunc: func() (string, context.Context) {
				// This test needs a client that blocks private networks, so we'll handle this specially
				return "SSRF_TEST_SPECIAL", context.Background()
			},
			expectedCode: ErrSSRF,
			shouldError:  true,
		},
		{
			name: "ErrFetch - network error",
			setupFunc: func() (string, context.Context) {
				// Use a non-existent domain to trigger network error
				return "http://thisdoesnotexist.invalid/test", context.Background()
			},
			expectedCode: ErrFetch,
			shouldError:  true,
		},
		{
			name: "ErrFetch - connection refused",
			setupFunc: func() (string, context.Context) {
				// Use a local port that's not listening
				return "http://localhost:99999/test", context.Background()
			},
			expectedCode: ErrFetch,
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, ctx := tt.setupFunc()
			
			// Handle SSRF test specially - needs a client that blocks private networks
			testClient := client
			if url == "SSRF_TEST_SPECIAL" {
				testClient = New() // Default client blocks private networks
				url = "http://192.168.1.1/test" // Use private IP to trigger SSRF
			}
			
			result, err := testClient.Parse(ctx, url)
			
			if !tt.shouldError {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result, got nil")
				}
				return
			}

			// Should have error
			if err == nil {
				t.Fatal("Expected error, got none")
			}

			// Check error type
			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("Expected ParseError, got: %T", err)
			}

			if parseErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %d (%s), got %d (%s). Error: %v", 
					tt.expectedCode, ErrorCode(tt.expectedCode).String(), 
					parseErr.Code, parseErr.Code.String(), 
					parseErr.Error())
			}

			if result != nil {
				t.Errorf("Expected nil result on error, got: %v", result)
			}
		})
	}
}

// TestParseErrorMethods tests the ParseError interface methods
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

// TestErrorWrappingAndUnwrapping tests error wrapping behavior
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

// TestParseHTMLErrorHandling tests error handling in ParseHTML
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
			
			if err == nil {
				t.Fatal("Expected error, got none")
			}

			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("Expected ParseError, got: %T", err)
			}

			if parseErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %d, got %d", tt.expectedCode, parseErr.Code)
			}

			if parseErr.Op != "ParseHTML" {
				t.Errorf("Expected operation 'ParseHTML', got '%s'", parseErr.Op)
			}

			if result != nil {
				t.Error("Expected nil result on error")
			}
		})
	}
}

// TestContextCancellationErrorClassification tests that context cancellation is properly classified
func TestContextCancellationErrorClassification(t *testing.T) {
	// Create a server that responds slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte("<html><body>Test content</body></html>"))
	}))
	defer server.Close()

	client := New(WithAllowPrivateNetworks(true)) // Allow localhost for testing

	t.Run("deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		result, err := client.Parse(ctx, server.URL)

		if err == nil {
			t.Fatal("Expected timeout error, got none")
		}

		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Fatalf("Expected ParseError, got: %T", err)
		}

		if parseErr.Code != ErrTimeout {
			t.Errorf("Expected ErrTimeout, got %d", parseErr.Code)
		}

		if result != nil {
			t.Error("Expected nil result on timeout")
		}
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		
		// Cancel the context immediately
		cancel()

		result, err := client.Parse(ctx, server.URL)

		if err == nil {
			t.Fatal("Expected cancellation error, got none")
		}

		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Fatalf("Expected ParseError, got: %T", err)
		}

		// Context cancellation should be classified as timeout
		if parseErr.Code != ErrTimeout {
			t.Errorf("Expected ErrTimeout for cancellation, got %d", parseErr.Code)
		}

		if result != nil {
			t.Error("Expected nil result on cancellation")
		}
	})
}

// TestNetworkErrorClassification tests classification of various network errors
func TestNetworkErrorClassification(t *testing.T) {
	client := New(WithAllowPrivateNetworks(true)) // Allow localhost for testing
	ctx := context.Background()

	tests := []struct {
		name         string
		url          string
		expectedCode ErrorCode
		description  string
	}{
		{
			name:         "DNS resolution failure",
			url:          "http://definitely-does-not-exist.invalid",
			expectedCode: ErrFetch,
			description:  "non-existent domain should trigger DNS error",
		},
		{
			name:         "Connection refused",
			url:          "http://localhost:99999",
			expectedCode: ErrFetch,
			description:  "unused port should trigger connection refused",
		},
		// Note: Private network test moved to separate function since this client allows private networks
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.Parse(ctx, tt.url)

			if err == nil {
				t.Fatalf("Expected error for %s, got none", tt.description)
			}

			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("Expected ParseError, got: %T", err)
			}

			if parseErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %d for %s, got %d", tt.expectedCode, tt.description, parseErr.Code)
			}

			if result != nil {
				t.Error("Expected nil result on error")
			}
		})
	}
}

// TestSSRFProtectionNetworkErrors tests that SSRF protection properly blocks private networks
func TestSSRFProtectionNetworkErrors(t *testing.T) {
	// Use default client (no private networks allowed)
	client := New()
	ctx := context.Background()

	tests := []struct {
		name         string
		url          string
		expectedCode ErrorCode
		description  string
	}{
		{
			name:         "Private network blocked",
			url:          "http://192.168.1.1",
			expectedCode: ErrSSRF,
			description:  "private IP should be blocked by SSRF protection",
		},
		{
			name:         "Localhost blocked",
			url:          "http://127.0.0.1:8080",
			expectedCode: ErrSSRF,
			description:  "localhost should be blocked by SSRF protection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.Parse(ctx, tt.url)

			if err == nil {
				t.Fatalf("Expected error for %s, got none", tt.description)
			}

			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("Expected ParseError, got: %T", err)
			}

			if parseErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %d for %s, got %d. Error: %v", tt.expectedCode, tt.description, parseErr.Code, parseErr.Error())
			}

			if result != nil {
				t.Error("Expected nil result on error")
			}
		})
	}
}

// TestErrorCodeValues tests that error codes have expected values
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

// BenchmarkErrorClassification benchmarks the error classification performance
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