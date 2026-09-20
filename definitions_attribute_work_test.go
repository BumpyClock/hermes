package hermes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDefinitionsAttributeWriteChargesKeyBytes(t *testing.T) {
	support := DefinitionCapabilities()
	longName := strings.Repeat("a", support.MaxStringBytes)
	source := "<article>" + strings.Repeat("<p>Article value.</p>", support.MaxTransformWork/(support.MaxStringBytes+1)+2) + "</article>"
	definitions := transformDefinitions(t, `    - target: descendants
      selector: p
      attribute.set: {name: `+longName+`, value: ""}
`)
	const articleURL = "https://93.184.216.34/story"
	client := New(WithDefinitions(definitions), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
	})))
	for _, op := range []string{"Parse", "ParseHTML"} {
		var result *Result
		var err error
		if op == "Parse" {
			result, err = client.Parse(context.Background(), articleURL)
		} else {
			result, err = client.ParseHTML(context.Background(), source, articleURL)
		}
		var parseError *ParseError
		if result != nil || !errors.Is(err, ErrDefinitionTransformLimit) || !errors.As(err, &parseError) || parseError.Code != ErrExtract || parseError.Op != op {
			t.Fatalf("%s did not charge generated attribute keys: result=%v error=%v", op, result, err)
		}
	}
}

func TestDefinitionsSmallAttributeWritesRemainValid(t *testing.T) {
	definitions := transformDefinitions(t, `    - target: descendants
      selector: p
      attribute.set: {name: data-safe, value: ""}
    - target: descendants
      selector: p
      element.rename: {tag: blockquote}
`)
	source := "<article>" + strings.Repeat("<p>Small attribute value.</p>", 256) + "</article>"
	r, err := New(WithDefinitions(definitions)).ParseHTML(context.Background(), source, "https://93.184.216.34/story")
	if err != nil || strings.Count(r.Content, "<blockquote>") != 256 {
		t.Fatalf("small bounded attributes no longer pass: result=%v error=%v", r, err)
	}
}
