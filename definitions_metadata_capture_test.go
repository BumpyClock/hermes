package hermes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/BumpyClock/hermes/internal/extractors"
)

func TestDefinitionsMetadataTextCaptureDateFallback(t *testing.T) {
	s, _ := localDefinitions(t, `schema: 1
site: metadata-capture
hosts: [93.184.216.34]
metadata:
  title: [{text: h1}]
  author: [{text: ".author a:first-child"}]
  date_published:
    - text_capture:
        selector: .author
        pattern: 'Published ([^ ]+)'
        group: 1
    - text_capture:
        selector: .author
        pattern: '([0-9]{4}-[0-9]{2}-[0-9]{2})'
        group: 1
content:
  groups: [[article]]
  default_cleaner: false
`)
	source := `<h1>Synthetic captured metadata report</h1><div class="author"><a href="/author">Ada Reporter</a> Published not-a-date · 2026-09-17</div><article><p>Captured metadata article opening.</p><p>Captured metadata article closing.</p></article>`
	const articleURL = "https://93.184.216.34/story"
	for _, format := range []string{"html", "markdown", "text"} {
		for _, op := range []string{"Parse", "ParseHTML"} {
			t.Run(format+"/"+op, func(t *testing.T) {
				c := New(WithDefinitions(s), WithContentType(format), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
				})))
				var result *Result
				var err error
				if op == "Parse" {
					result, err = c.Parse(context.Background(), articleURL)
				} else {
					result, err = c.ParseHTML(context.Background(), source, articleURL)
				}
				if err != nil {
					t.Fatal(err)
				}
				if result.Author != "Ada Reporter" || result.DatePublished == nil || result.DatePublished.Format("2006-01-02") != "2026-09-17" {
					t.Fatalf("captured metadata: %+v", result)
				}
				if !strings.Contains(result.Content, "Captured metadata article opening.") {
					t.Fatalf("%s content lost: %s", format, result.Content)
				}
			})
		}
	}
}

func TestDefinitionsMetadataTextCaptureLimitIsPublic(t *testing.T) {
	s, _ := localDefinitions(t, `schema: 1
site: metadata-limit
hosts: [93.184.216.34]
metadata:
  date_published:
    - text_capture:
        selector: .author
        pattern: '([0-9]{4}-[0-9]{2}-[0-9]{2})'
        group: 1
content:
  groups: [[article]]
`)
	source := `<div class="author">` + strings.Repeat("a", DefinitionCapabilities().MaxValueBytes+1) + `</div><article><p>Metadata limits stay observable.</p></article>`
	const articleURL = "https://93.184.216.34/story"
	c := New(WithDefinitions(s), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
	})))
	for _, op := range []string{"Parse", "ParseHTML"} {
		var result *Result
		var err error
		if op == "Parse" {
			result, err = c.Parse(context.Background(), articleURL)
		} else {
			result, err = c.ParseHTML(context.Background(), source, articleURL)
		}
		var parseError *ParseError
		var captureError *extractors.MetadataCaptureError
		if result != nil || !errors.As(err, &parseError) || parseError.Code != ErrExtract || parseError.Op != op || !errors.As(err, &captureError) || !errors.Is(err, extractors.ErrMetadataCaptureLimit) {
			t.Fatalf("%s did not preserve metadata capture limit: result=%v error=%v", op, result, err)
		}
	}
}

type metadataCaptureContext struct {
	context.Context
	calls, cancelAt int
}

func (c *metadataCaptureContext) Err() error {
	c.calls++
	if c.calls >= c.cancelAt {
		return context.Canceled
	}
	return nil
}

func TestDefinitionsMetadataTextCaptureCancellation(t *testing.T) {
	s, _ := localDefinitions(t, `schema: 1
site: metadata-cancellation
hosts: [93.184.216.34]
metadata:
  date_published:
    - text_capture:
        selector: .author
        pattern: '([0-9]{4}-[0-9]{2}-[0-9]{2})'
        group: 1
content:
  groups: [[article]]
`)
	source := `<div class="author">` + strings.Repeat("<span>metadata </span>", 128) + `2026-09-17</div><article><p>Cancellation stays observable.</p></article>`
	for _, op := range []string{"Parse", "ParseHTML"} {
		ctx := &metadataCaptureContext{Context: context.Background(), cancelAt: 24}
		c := New(WithDefinitions(s), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
		})))
		var result *Result
		var err error
		if op == "Parse" {
			result, err = c.Parse(ctx, "https://93.184.216.34/story")
		} else {
			result, err = c.ParseHTML(ctx, source, "https://93.184.216.34/story")
		}
		var parseError *ParseError
		if result != nil || !errors.Is(err, context.Canceled) || !errors.As(err, &parseError) || parseError.Code != ErrTimeout || parseError.Op != op {
			t.Fatalf("%s did not preserve cancellation: result=%v error=%v calls=%d", op, result, err, ctx.calls)
		}
	}
}
