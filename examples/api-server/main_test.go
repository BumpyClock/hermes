package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/BumpyClock/hermes"
)

type mockParser struct {
	content string
	urls    []string
}

func (p *mockParser) Parse(_ context.Context, targetURL string) (*hermes.Result, error) {
	p.urls = append(p.urls, targetURL)
	return &hermes.Result{URL: targetURL, Content: p.content}, nil
}

func (p *mockParser) ParseHTML(ctx context.Context, _ string, targetURL string) (*hermes.Result, error) {
	return p.Parse(ctx, targetURL)
}

func TestParseFormats(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		for _, test := range []struct {
			format      string
			content     string
			contentType string
			json        bool
		}{
			{"", "html", "application/json", true},
			{"json", "html", "application/json", true},
			{"html", "html", "text/html; charset=utf-8", false},
			{"markdown", "markdown", "text/markdown; charset=utf-8", false},
			{"text", "text", "text/plain; charset=utf-8", false},
			{"JSON", "html", "; charset=utf-8", false},
			{"Html", "html", "; charset=utf-8", false},
			{"MarkDown", "html", "; charset=utf-8", false},
			{"TEXT", "html", "; charset=utf-8", false},
		} {
			t.Run(method+"/"+test.format, func(t *testing.T) {
				parsers := map[string]*mockParser{
					"html": {content: "html"}, "markdown": {content: "markdown"}, "text": {content: "text"},
				}
				server := &Server{clients: map[string]hermes.Parser{
					"html": parsers["html"], "markdown": parsers["markdown"], "text": parsers["text"],
				}}
				targetURL := "https://example.com/article"
				request := parseRequest(t, method, targetURL, test.format)
				response := httptest.NewRecorder()
				server.handleParse(response, request)
				if response.Code != http.StatusOK {
					t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
				}
				if got := response.Header().Get("Content-Type"); got != test.contentType {
					t.Fatalf("content type = %q, want %q", got, test.contentType)
				}
				if test.json {
					var body ParseResponse
					if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
						t.Fatal(err)
					}
					if !body.Success || body.Data == nil || body.Data.Content != test.content || body.Metadata == nil {
						t.Fatalf("JSON response = %+v", body)
					}
				} else if response.Body.String() != test.content {
					t.Fatalf("body = %q, want %q", response.Body.String(), test.content)
				}
				for name, parser := range parsers {
					wantCalls := 0
					if name == test.content {
						wantCalls = 1
					}
					if len(parser.urls) != wantCalls || wantCalls == 1 && parser.urls[0] != targetURL {
						t.Fatalf("%s parser URLs = %v", name, parser.urls)
					}
				}
			})
		}
	}
}

func TestParseRejectsUnsupportedFormats(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		for _, format := range []string{"md", "txt", "text/html", " json ", "xml"} {
			t.Run(method+"/"+format, func(t *testing.T) {
				response := httptest.NewRecorder()
				(&Server{}).handleParse(response, parseRequest(t, method, "https://example.com/article", format))
				if response.Code != http.StatusBadRequest {
					t.Fatalf("status = %d, want 400", response.Code)
				}
				var body ParseResponse
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Error == nil || body.Error.Code != "invalid_format" {
					t.Fatalf("error = %+v, want invalid_format", body.Error)
				}
			})
		}
	}
}

func parseRequest(t *testing.T, method, targetURL, format string) *http.Request {
	t.Helper()
	if method == http.MethodGet {
		query := url.Values{"url": {targetURL}, "format": {format}}
		return httptest.NewRequest(method, "/parse?"+query.Encode(), nil)
	}
	body, err := json.Marshal(ParseRequest{URL: targetURL, Format: format})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(method, "/parse", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	return request
}
