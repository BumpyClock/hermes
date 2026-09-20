package hermes

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestUnconfiguredClientsAreGenericOnly(t *testing.T) {
	const source = `<html><head><title>Generic-only headline</title></head><body><article><p>Original synthetic article content remains available without configured definitions.</p><p>The closing generic paragraph supplies enough context for extraction.</p></article></body></html>`
	const articleURL = "https://93.184.216.34/article"
	for _, configured := range []bool{false, true} {
		for _, op := range []string{"Parse", "ParseHTML"} {
			t.Run(strings.Join([]string{map[bool]string{false: "unconfigured", true: "empty snapshot"}[configured], op}, "/"), func(t *testing.T) {
				options := []Option{WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
				}))}
				if configured {
					options = append(options, WithDefinitions(nil))
				}
				client := New(options...)
				var result *Result
				var err error
				if op == "Parse" {
					result, err = client.Parse(context.Background(), articleURL)
				} else {
					result, err = client.ParseHTML(context.Background(), source, articleURL)
				}
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(result.Content, "Original synthetic article content") {
					t.Fatalf("expected generic-only result, got %+v", result)
				}
			})
		}
	}
}
