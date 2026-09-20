package hermes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestDefinitionsAttributeScansChargeOptionalMissesAndRemovals(t *testing.T) {
	var steps strings.Builder
	for i := 0; i < 16; i++ {
		steps.WriteString(`    - target: root
      attribute.copy: {from: missing, to: data-copy-` + strconv.Itoa(i) + `}
    - target: root
      attribute.remove: {name: data-remove-` + strconv.Itoa(i) + `}
`)
	}
	definitions := transformDefinitions(t, steps.String())
	source := attributeHeavyArticle(512, 128)
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
			t.Fatalf("%s did not charge inspected attributes: result=%v error=%v", op, result, err)
		}
	}
}

type attributeScanContext struct {
	context.Context
	calls, cancelAt int
}

func (c *attributeScanContext) Err() error {
	c.calls++
	if c.calls >= c.cancelAt {
		return context.Canceled
	}
	return nil
}

func TestDefinitionsAttributeScanCancellation(t *testing.T) {
	definitions := transformDefinitions(t, `    - target: root
      attribute.copy: {from: missing, to: data-copy}
`)
	source := attributeHeavyArticle(256, 64)
	const articleURL = "https://93.184.216.34/story"
	for _, op := range []string{"Parse", "ParseHTML"} {
		ctx := &attributeScanContext{Context: context.Background(), cancelAt: 300}
		client := New(WithDefinitions(definitions), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
		})))
		var result *Result
		var err error
		if op == "Parse" {
			result, err = client.Parse(ctx, articleURL)
		} else {
			result, err = client.ParseHTML(ctx, source, articleURL)
		}
		var parseError *ParseError
		var operationError *DefinitionOperationError
		if result != nil || !errors.Is(err, context.Canceled) || !errors.As(err, &parseError) || parseError.Code != ErrTimeout || parseError.Op != op || !errors.As(err, &operationError) || operationError.Step != 1 {
			t.Fatalf("%s did not preserve scan cancellation: result=%v error=%v calls=%d", op, result, err, ctx.calls)
		}
	}
}

func attributeHeavyArticle(attributes, valueBytes int) string {
	value := strings.Repeat("v", valueBytes)
	var source strings.Builder
	source.WriteString("<article")
	for i := 0; i < attributes; i++ {
		source.WriteString(` data-source-` + strconv.Itoa(i) + `="` + value + `"`)
	}
	source.WriteString(">Attribute-heavy article.</article>")
	return source.String()
}
