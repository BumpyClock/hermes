package hermes

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDefinitionsTransformLiteralSpaceReplacement(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: img
      string.replace: {attribute: alt, old: ' ', new: '-'}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><p>Report.</p><img src="/photo.jpg" alt="Two words"></article>`, "https://93.184.216.34/story")
	if err != nil || !strings.Contains(r.Content, `alt="Two-words"`) {
		t.Fatalf("literal whitespace replacement: %+v, %v", r, err)
	}
}

func TestDefinitionsTransformRequiredEmptyCapture(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: img
      regex.capture: {attribute: alt, pattern: 'value:(.*)', group: 1, required: true}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><p>Report.</p><img src="/photo.jpg" alt="value:"></article>`, "https://93.184.216.34/story")
	var pe *ParseError
	if r != nil || !errors.As(err, &pe) || pe.Code != ErrExtract {
		t.Fatalf("empty required capture silently accepted: %+v, %v", r, err)
	}
}

func TestDefinitionsTransformStepDoesNotRequeryNewMatches(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: "div:not(:has(div))"
      element.unwrap: {}
    - target: root
      when: [{descendant: {selector: div}}]
      element.rename: {tag: blockquote}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><div><div><p>Inner content stays visible.</p></div></div></article>`, "https://93.184.216.34/story")
	if err != nil || !strings.Contains(r.Content, "<blockquote>") || !strings.Contains(r.Content, "Inner content") {
		t.Fatalf("new outer match repeatedly consumed: %+v, %v", r, err)
	}
}

type transformCancelContext struct {
	context.Context
	calls int
	cause error
}

func (c *transformCancelContext) Err() error {
	c.calls++
	if c.calls > 700 {
		return c.cause
	}
	return nil
}

func TestDefinitionsTransformPublicCancellation(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: p
      attribute.set: {name: title, value: Changed}
`)
	source := "<article>" + strings.Repeat("<p>Original</p>", 100) + "</article>"
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		ctx := &transformCancelContext{Context: context.Background(), cause: cause}
		r, err := New(WithDefinitions(s)).ParseHTML(ctx, source, "https://93.184.216.34/story")
		var pe *ParseError
		var de *DefinitionOperationError
		if r != nil || !errors.Is(err, cause) || !errors.As(err, &pe) || pe.Code != ErrTimeout || pe.Op != "ParseHTML" || !errors.As(err, &de) {
			t.Fatalf("lost in-executor cancellation classification: %+v, %v", r, err)
		}
	}
}

func TestDefinitionsTransformValueBoundary(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: img
      string.replace: {attribute: alt, old: a, new: b}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><p>Report.</p><img src="/photo.jpg" alt="`+strings.Repeat("a", DefinitionCapabilities().MaxValueBytes)+`"></article>`, "https://93.184.216.34/story")
	if err != nil || !strings.Contains(r.Content, strings.Repeat("b", DefinitionCapabilities().MaxValueBytes)) {
		t.Fatalf("maximum valid scalar rejected: %v", err)
	}
}
