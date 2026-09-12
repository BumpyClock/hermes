package parser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestParserReusesDefaultHTTPConnections(t *testing.T) {
	for _, explicitOptions := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit_options=%t", explicitOptions), func(t *testing.T) {
			peers := make(chan string, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				peers <- r.RemoteAddr
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte(sampleNewsHTML))
			}))
			defer server.Close()

			options := DefaultParserOptions()
			options.AllowPrivateNetworks = true
			parser := New(options)
			var requestOptions *ParserOptions
			if explicitOptions {
				requestOptions = options
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			for range 2 {
				result, err := parser.ParseWithContext(ctx, server.URL, requestOptions)
				if err != nil {
					t.Fatal(err)
				}
				if result.Content == "" {
					t.Fatal("expected article content")
				}
			}
			if first, second := <-peers, <-peers; first != second {
				t.Fatalf("connection was not reused: %s then %s", first, second)
			}
			if options.HTTPClient != nil {
				t.Fatal("parser mutated caller options")
			}
		})
	}
}

func TestParserConcurrentDefaultHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleNewsHTML))
	}))
	defer server.Close()

	options := DefaultParserOptions()
	options.AllowPrivateNetworks = true
	parser := New(options)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			result, err := parser.ParseWithContext(ctx, server.URL, nil)
			if err != nil {
				t.Error(err)
				return
			}
			if result.Content == "" {
				t.Error("expected article content")
			}
		}()
	}
	workers.Wait()
}

func TestParserUsesProvidedHTTPClient(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleNewsHTML))
	}))
	defer server.Close()

	for _, explicitOptions := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit_options=%t", explicitOptions), func(t *testing.T) {
			options := DefaultParserOptions()
			options.AllowPrivateNetworks = true
			options.HTTPClient = server.Client()
			parser := New(options)
			var requestOptions *ParserOptions
			if explicitOptions {
				parser = New()
				requestOptions = options
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			result, err := parser.ParseWithContext(ctx, server.URL, requestOptions)
			if err != nil {
				t.Fatal(err)
			}
			if result.Content == "" {
				t.Fatal("expected article content")
			}
		})
	}
}

func TestParserProvidedHTTPClientOverridesInitializedDefault(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleNewsHTML))
	}))
	defer httpServer.Close()

	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleNewsHTML))
	}))
	defer tlsServer.Close()

	options := DefaultParserOptions()
	options.AllowPrivateNetworks = true
	parser := New(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := parser.ParseWithContext(ctx, httpServer.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Content == "" {
		t.Fatal("expected article content")
	}

	override := DefaultParserOptions()
	override.AllowPrivateNetworks = true
	override.HTTPClient = tlsServer.Client()
	result, err = parser.ParseWithContext(ctx, tlsServer.URL, override)
	if err != nil {
		t.Fatal(err)
	}
	if result.Content == "" {
		t.Fatal("expected article content")
	}

	result, err = parser.ParseWithContext(ctx, httpServer.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Content == "" {
		t.Fatal("expected article content")
	}
}

func TestParserDefaultHTTPClientRequestHeadersDoNotLeak(t *testing.T) {
	got := make(chan http.Header, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- r.Header.Clone()
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleNewsHTML))
	}))
	defer server.Close()

	options := DefaultParserOptions()
	options.AllowPrivateNetworks = true
	parser := New(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	firstOpts := DefaultParserOptions()
	firstOpts.AllowPrivateNetworks = true
	firstOpts.Headers["X-Test-Header"] = "first"
	result, err := parser.ParseWithContext(ctx, server.URL, firstOpts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Content == "" {
		t.Fatal("expected article content")
	}

	secondOpts := DefaultParserOptions()
	secondOpts.AllowPrivateNetworks = true
	secondOpts.Headers["X-Test-Header"] = "second"
	result, err = parser.ParseWithContext(ctx, server.URL, secondOpts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Content == "" {
		t.Fatal("expected article content")
	}

	result, err = parser.ParseWithContext(ctx, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Content == "" {
		t.Fatal("expected article content")
	}

	first := <-got
	if first.Get("X-Test-Header") != "first" {
		t.Fatalf("first request header = %q, want first", first.Get("X-Test-Header"))
	}
	second := <-got
	if second.Get("X-Test-Header") != "second" {
		t.Fatalf("second request header = %q, want second", second.Get("X-Test-Header"))
	}
	third := <-got
	if leaked := third.Get("X-Test-Header"); leaked != "" {
		t.Fatalf("header leaked onto later request: %q", leaked)
	}
	if firstOpts.Headers["X-Test-Header"] != "first" || secondOpts.Headers["X-Test-Header"] != "second" {
		t.Fatal("parser mutated caller headers")
	}
}
