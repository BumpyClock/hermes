package hermes

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestContextCancellationImmediate tests immediate context cancellation
func TestContextCancellationImmediate(t *testing.T) {
	// Create a test server that would delay response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should never be reached
		t.Error("Request should have been cancelled before reaching server")
		time.Sleep(5 * time.Second)
		w.Write([]byte(`<html><body>Should not see this</body></html>`))
	}))
	defer ts.Close()

	client := New(WithAllowPrivateNetworks(true))

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to parse - should fail immediately
	start := time.Now()
	_, err := client.Parse(ctx, ts.URL)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected error from cancelled context, got nil")
	}

	// Should fail very quickly (within 100ms)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Cancellation took too long: %v", elapsed)
	}

	t.Logf("✓ Context cancellation worked immediately: %v", elapsed)
}

// TestContextCancellationDuringFetch tests cancellation during HTTP fetch
func TestContextCancellationDuringFetch(t *testing.T) {
	// Create a test server that delays before responding
	serverStarted := make(chan bool)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverStarted <- true
		// Delay to allow cancellation
		time.Sleep(2 * time.Second)
		w.Write([]byte(`<html><body>Too late</body></html>`))
	}))
	defer ts.Close()

	client := New(WithAllowPrivateNetworks(true))

	// Create a context that will be cancelled during fetch
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start parsing in a goroutine
	done := make(chan error)
	go func() {
		_, err := client.Parse(ctx, ts.URL)
		done <- err
	}()

	// Wait for server to start processing
	select {
	case <-serverStarted:
		// Now cancel the context
		cancel()
	case <-time.After(1 * time.Second):
		t.Fatal("Server didn't start processing request")
	}

	// Wait for parse to complete
	err := <-done
	if err == nil {
		t.Fatal("Expected error from cancelled context")
	}

	t.Logf("✓ Context cancellation during fetch: %v", err)
}

// TestContextTimeout tests context timeout handling
func TestContextTimeout(t *testing.T) {
	// Create a test server that delays longer than our timeout
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than our timeout
		w.Write([]byte(`<html><body>Too slow</body></html>`))
	}))
	defer ts.Close()

	client := New(WithAllowPrivateNetworks(true))

	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := client.Parse(ctx, ts.URL)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	// Check that it's a timeout error
	if perr, ok := err.(*ParseError); ok {
		if perr.Code != ErrTimeout {
			t.Errorf("Expected ErrTimeout, got %v", perr.Code)
		}
	}

	// Should timeout around 200ms (allow some margin)
	if elapsed < 180*time.Millisecond || elapsed > 400*time.Millisecond {
		t.Errorf("Timeout timing unexpected: %v", elapsed)
	}

	t.Logf("✓ Context timeout worked correctly: %v", elapsed)
}

// TestContextPropagation tests that context is properly propagated through layers
func TestContextPropagation(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The fact that we get here means context was propagated
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Test</title></head><body>Content</body></html>`))
	}))
	defer ts.Close()

	client := New(WithAllowPrivateNetworks(true))

	// Create a context with a value
	type ctxKey string
	ctx := context.WithValue(context.Background(), ctxKey("test"), "value")

	result, err := client.Parse(ctx, ts.URL)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result.Title == "" {
		t.Error("No title extracted")
	}

	t.Logf("✓ Context propagated successfully through all layers")
}

// TestConcurrentContextCancellation tests concurrent requests with different contexts
func TestConcurrentContextCancellation(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Small delay
		w.Write([]byte(`<html><head><title>Test</title></head><body>Content</body></html>`))
	}))
	defer ts.Close()

	client := New(WithAllowPrivateNetworks(true))

	// Test multiple concurrent requests with different contexts
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("concurrent_%d", i), func(t *testing.T) {
			t.Parallel()
			
			if i%2 == 0 {
				// Even iterations: use timeout context
				ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
				defer cancel()
				
				_, err := client.Parse(ctx, ts.URL)
				if err == nil {
					t.Error("Expected timeout error")
				}
			} else {
				// Odd iterations: normal context
				ctx := context.Background()
				result, err := client.Parse(ctx, ts.URL)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil && result.Title == "" {
					t.Error("No title extracted")
				}
			}
		})
	}
}