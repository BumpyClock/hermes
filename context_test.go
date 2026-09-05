package hermes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestContextCancellationImmediate(t *testing.T) {
	client := New(WithTransport(testTransport(func(*http.Request) (*http.Response, error) {
		t.Error("canceled request reached transport")
		return nil, context.Canceled
	})))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Parse(ctx, "https://8.8.8.8/article")
	if !errors.Is(err, &ParseError{Code: ErrTimeout}) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestContextCancellationDuringFetch(t *testing.T) {
	started := make(chan struct{})
	client := New(WithTransport(testTransport(func(req *http.Request) (*http.Response, error) {
		close(started)
		<-req.Context().Done()
		return nil, req.Context().Err()
	})))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := client.Parse(ctx, "https://8.8.8.8/article")
		done <- err
	}()
	select {
	case <-started:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("request did not reach transport")
	}
	select {
	case err := <-done:
		if !errors.Is(err, &ParseError{Code: ErrTimeout}) {
			t.Fatalf("expected context cancellation, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("request did not return after cancellation")
	}
}

func TestContextTimeout(t *testing.T) {
	client := New(WithTransport(testTransport(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := client.Parse(ctx, "https://8.8.8.8/article")
	var parseErr *ParseError
	if !errors.As(err, &parseErr) || parseErr.Code != ErrTimeout {
		t.Fatalf("expected deadline error with ErrTimeout, got %v", err)
	}
}

func TestContextPropagation(t *testing.T) {
	type contextKey struct{}
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), contextKey{}, "value"), 5*time.Second)
	defer cancel()
	deadline, _ := ctx.Deadline()
	called := false
	client := New(WithTransport(testTransport(func(req *http.Request) (*http.Response, error) {
		called = true
		if got := req.Context().Value(contextKey{}); got != "value" {
			t.Errorf("context value = %v, want value", got)
		}
		if got, ok := req.Context().Deadline(); !ok || !got.Equal(deadline) {
			t.Errorf("context deadline = %v (%v), want %v", got, ok, deadline)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"text/html"}},
			Body:       io.NopCloser(strings.NewReader("<html><head><title>Test</title></head><body><article>Context test content.</article></body></html>")),
			Request:    req,
		}, nil
	})))
	result, err := client.Parse(ctx, "https://8.8.8.8/article")
	if err != nil {
		t.Fatal(err)
	}
	if !called || result.Title != "Test" {
		t.Fatalf("transport called = %v, title = %q", called, result.Title)
	}
}

func TestConcurrentContextCancellation(t *testing.T) {
	started := make(chan struct{})
	client := New(WithTransport(testTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/cancel" {
			close(started)
			<-req.Context().Done()
			return nil, req.Context().Err()
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"text/html"}},
			Body:       io.NopCloser(strings.NewReader("<html><head><title>Unaffected</title></head><body>Content</body></html>")),
			Request:    req,
		}, nil
	})))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := client.Parse(ctx, "https://8.8.8.8/cancel")
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("request did not reach transport")
	}
	cancel()
	result, err := client.Parse(context.Background(), "https://8.8.8.8/success")
	if err != nil {
		t.Fatal(err)
	}
	if result.Title != "Unaffected" {
		t.Errorf("title = %q, want Unaffected", result.Title)
	}
	select {
	case err := <-done:
		if !errors.Is(err, &ParseError{Code: ErrTimeout}) {
			t.Fatalf("expected cancellation, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("canceled request did not return")
	}
}
