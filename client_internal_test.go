package hermes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewWithTimeoutPreservesDefaultTransport(t *testing.T) {
	client := New(WithTimeout(5 * time.Second))

	if client.httpClient == nil {
		t.Fatal("expected HTTP client")
	}
	if client.httpClient.Timeout != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", client.httpClient.Timeout)
	}

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected default *http.Transport, got %T", client.httpClient.Transport)
	}
	if transport.MaxIdleConns != 100 {
		t.Fatalf("expected MaxIdleConns 100, got %d", transport.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != 10 {
		t.Fatalf("expected MaxIdleConnsPerHost 10, got %d", transport.MaxIdleConnsPerHost)
	}
}

func TestNewUsesCurrentChromeUserAgentByDefault(t *testing.T) {
	client := New()

	if client.userAgent != DefaultUserAgent {
		t.Fatalf("expected default User-Agent %q, got %q", DefaultUserAgent, client.userAgent)
	}
}

func TestWithUserAgentSupportsFacebookCrawler(t *testing.T) {
	client := New(WithUserAgent(FacebookCrawlerUserAgent))

	if client.userAgent != FacebookCrawlerUserAgent {
		t.Fatalf("expected Facebook crawler User-Agent %q, got %q", FacebookCrawlerUserAgent, client.userAgent)
	}
}

func TestParseSendsConfiguredUserAgent(t *testing.T) {
	tests := []struct {
		name       string
		options    []Option
		wantHeader string
	}{
		{
			name:       "default",
			wantHeader: DefaultUserAgent,
		},
		{
			name:       "Facebook crawler",
			options:    []Option{WithUserAgent(FacebookCrawlerUserAgent)},
			wantHeader: FacebookCrawlerUserAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("User-Agent"); got != tt.wantHeader {
					t.Errorf("expected User-Agent %q, got %q", tt.wantHeader, got)
				}
				w.Header().Set("Content-Type", "text/html")
				_, _ = w.Write([]byte("<html><head><title>Test article</title></head><body><article><p>Content</p></article></body></html>"))
			}))
			defer server.Close()

			opts := append([]Option{WithAllowPrivateNetworks(true)}, tt.options...)
			if _, err := New(opts...).Parse(context.Background(), server.URL); err != nil {
				t.Fatalf("Parse returned an error: %v", err)
			}
		})
	}
}

func TestNewWithTransportAppliesDefaultTimeout(t *testing.T) {
	transport := http.DefaultTransport
	client := New(WithTransport(transport))

	if client.httpClient.Transport != transport {
		t.Fatalf("expected custom transport to be preserved")
	}
	if client.httpClient.Timeout != 30*time.Second {
		t.Fatalf("expected default timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestNewWithHTTPClientTimeoutComposition(t *testing.T) {
	provided := &http.Client{}
	client := New(WithHTTPClient(provided))

	if client.httpClient != provided {
		t.Fatalf("expected provided HTTP client to be preserved")
	}
	if provided.Timeout != 0 {
		t.Fatalf("expected provided client timeout unchanged, got %v", provided.Timeout)
	}

	providedWithTimeout := &http.Client{}
	client = New(WithHTTPClient(providedWithTimeout), WithTimeout(7*time.Second))

	if client.httpClient != providedWithTimeout {
		t.Fatalf("expected provided HTTP client to be preserved")
	}
	if providedWithTimeout.Timeout != 7*time.Second {
		t.Fatalf("expected explicit timeout 7s on provided client, got %v", providedWithTimeout.Timeout)
	}
}
