package hermes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDefinitionsURLResolveAttributeBase(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: img[data-original]
      url.resolve:
        attribute: data-original
        to: src
        base: {attribute: src, required: true}
`)
	source := `<article><p>Wired-style image resolution.</p><img src="https://cdn.wired.jp/images/base/" data-original="relative.jpg" alt="Resolved"><p>Closing image resolution.</p></article>`
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
				if format != "text" && !strings.Contains(result.Content, "https://cdn.wired.jp/images/base/relative.jpg") {
					t.Fatalf("%s did not resolve attribute base: %s", format, result.Content)
				}
			})
		}
	}
}

func TestDefinitionsURLResolveRelativeTypedBase(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: img[data-original]
      url.resolve:
        attribute: data-original
        to: src
        base: {literal: "/images/base/"}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><p>Relative typed base.</p><img data-original="relative.jpg"></article>`, "https://93.184.216.34/story")
	if err != nil || !strings.Contains(r.Content, `src="https://93.184.216.34/images/base/relative.jpg"`) {
		t.Fatalf("relative typed base did not resolve against source URL: result=%v error=%v", r, err)
	}
}

func TestDefinitionsURLResolveOptionalAndUnsafeBase(t *testing.T) {
	optional := transformDefinitions(t, `    - target: descendants
      selector: img
      url.resolve:
        attribute: data-original
        to: src
        base: {attribute: missing}
    - target: descendants
      selector: img
      attribute.set: {name: alt, value: "Optional base skipped"}
`)
	r, err := New(WithDefinitions(optional)).ParseHTML(context.Background(), `<article><p>Optional base remains ordinary.</p><img src="/original.jpg" data-original="relative.jpg"></article>`, "https://93.184.216.34/story")
	if err != nil || !strings.Contains(r.Content, "original.jpg") || !strings.Contains(r.Content, "Optional base skipped") {
		t.Fatalf("optional base did not skip: result=%v error=%v", r, err)
	}

	unsafe := transformDefinitions(t, `    - target: descendants
      selector: img
      url.resolve:
        attribute: data-original
        to: src
        base: {attribute: src, required: true}
`)
	source := `<article><p>Unsafe base must fail.</p><img src="javascript:alert(1)" data-original="relative.jpg"></article>`
	c := New(WithDefinitions(unsafe), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
	})))
	for _, op := range []string{"Parse", "ParseHTML"} {
		var result *Result
		if op == "Parse" {
			result, err = c.Parse(context.Background(), "https://93.184.216.34/story")
		} else {
			result, err = c.ParseHTML(context.Background(), source, "https://93.184.216.34/story")
		}
		var parseError *ParseError
		if result != nil || !errors.As(err, &parseError) || parseError.Code != ErrExtract {
			t.Fatalf("%s unsafe base did not fail publicly: result=%v error=%v", op, result, err)
		}
	}
}
