package hermes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDefinitionsStructuredValuesCreateSafeMedia(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: .lazy-picture
      attribute.set_from:
        name: data-media-src
        value:
          json:
            attribute: data-props
            path: [{field: sources}, {index: 0}, {field: src}]
            required: true
    - target: descendants
      selector: .lazy-picture
      element.create:
        target: {selector: .slot, select: first}
        position: replace
        node:
          tag: img
          attributes:
            src: {attribute: data-media-src, required: true}
            alt: "Structured media"
    - target: descendants
      selector: .gallery
      element.create:
        target: {selector: .slot, select: first}
        position: append
        node:
          tag: img
          attributes:
            src:
              descendant_attribute:
                selector: .picturefill
                name: data-platform-src
                required: true
            alt: "Descendant media"
`)
	source := `<article><p>Structured media report.</p><div class="lazy-picture" data-props='{"sources":[{"src":"/apartment.jpg"}]}'><span class="slot">Placeholder</span></div><div class="gallery"><span class="picturefill" data-platform-src="/national.jpg"></span><span class="slot">Gallery slot</span></div><p>Closing structured media report.</p></article>`
	for _, format := range []string{"html", "markdown", "text"} {
		r, err := New(WithDefinitions(s), WithContentType(format)).ParseHTML(context.Background(), source, "https://93.184.216.34/story")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.Content, "Structured media report.") || !strings.Contains(r.Content, "Closing structured media report.") {
			t.Fatalf("%s lost article content: %s", format, r.Content)
		}
		if format != "text" {
			for _, image := range []string{"https://93.184.216.34/apartment.jpg", "https://93.184.216.34/national.jpg"} {
				if !strings.Contains(r.Content, image) {
					t.Errorf("%s missing structured image %q: %s", format, image, r.Content)
				}
			}
		}
	}
}

func TestDefinitionsCreateSelfReplacesStructuredWrapper(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: .lazy-picture
      attribute.set_from:
        name: data-media-src
        value:
          json:
            attribute: data-props
            path: [{field: sources}, {index: 0}, {field: src}]
            required: true
    - target: descendants
      selector: .lazy-picture
      element.create:
        target: {self: true}
        position: replace
        node:
          tag: img
          attributes:
            src: {attribute: data-media-src, required: true}
            alt: "Structured replacement"
`)
	source := `<article><p>Wrapper replacement report.</p><div class="lazy-picture" data-render-react-id="images/LazyPicture" data-props='{"sources":[{"src":"/replacement.jpg"}]}'></div></article>`
	const articleURL = "https://93.184.216.34/story"
	for _, op := range []string{"Parse", "ParseHTML"} {
		c := New(WithDefinitions(s), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
		})))
		var result *Result
		var err error
		if op == "Parse" {
			result, err = c.Parse(context.Background(), articleURL)
		} else {
			result, err = c.ParseHTML(context.Background(), source, articleURL)
		}
		if err != nil || !strings.Contains(result.Content, `src="https://93.184.216.34/replacement.jpg"`) || strings.Contains(result.Content, "data-render-react-id") || strings.Contains(result.Content, "lazy-picture") {
			t.Fatalf("%s did not replace the structured wrapper: result=%v error=%v", op, result, err)
		}
	}
}

func TestDefinitionsStructuredValueFailuresArePublic(t *testing.T) {
	tests := []struct {
		name, value, source, want string
	}{
		{"malformed JSON", `{json: {attribute: data-props, path: [{field: src}], required: true}}`, `data-props='{bad'`, "invalid character"},
		{"wrong JSON type", `{json: {attribute: data-props, path: [{field: src}], required: true}}`, `data-props='{"src":[]}'`, "must be a string"},
		{"index bounds", `{json: {attribute: data-props, path: [{field: sources}, {index: 2}, {field: src}], required: true}}`, `data-props='{"sources":[{"src":"/one.jpg"}]}'`, "out of bounds"},
		{"missing descendant attribute", `{descendant_attribute: {selector: img, name: src, required: true}}`, ``, "required descendant attribute"},
	}
	const articleURL = "https://93.184.216.34/story"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := transformDefinitions(t, `    - target: descendants
      selector: .media
      attribute.set_from:
        name: src
        value: `+test.value+`
`)
			source := `<article><p>Enough article text preserves an extraction failure.</p><div class="media" ` + test.source + `>Media</div></article>`
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
				var operationError *DefinitionOperationError
				if result != nil || !errors.As(err, &parseError) || parseError.Code != ErrExtract || parseError.Op != op || !errors.As(err, &operationError) || operationError.Step != 1 || !strings.Contains(err.Error(), test.want) {
					t.Fatalf("%s did not preserve structured failure context: result=%v error=%v", op, result, err)
				}
			}
		})
	}
}

func TestDefinitionsStructuredValueOptionalAndLimitBehavior(t *testing.T) {
	optional := transformDefinitions(t, `    - target: descendants
      selector: .media
      attribute.set_from:
        name: src
        value:
          json:
            attribute: data-props
            path: [{field: sources}, {index: 0}, {field: src}]
    - target: descendants
      selector: .media
      element.rename: {tag: blockquote}
`)
	r, err := New(WithDefinitions(optional)).ParseHTML(context.Background(), `<article><p>Optional structured input remains ordinary.</p><div class="media" data-props="{not-json}">Media</div></article>`, "https://93.184.216.34/story")
	if err != nil || !strings.Contains(r.Content, "<blockquote") {
		t.Fatalf("optional malformed JSON did not skip: result=%v error=%v", r, err)
	}

	deep := strings.Repeat(`{"a":`, DefinitionCapabilities().MaxJSONDepth+1) + `"/image.jpg"` + strings.Repeat(`}`, DefinitionCapabilities().MaxJSONDepth+1)
	limited := transformDefinitions(t, `    - target: descendants
      selector: .media
      attribute.set_from:
        name: src
        value:
          json:
            attribute: data-props
            path: [{field: a}]
            required: true
`)
	r, err = New(WithDefinitions(limited)).ParseHTML(context.Background(), `<article><p>Bounded JSON input.</p><div class="media" data-props='`+deep+`'>Media</div></article>`, "https://93.184.216.34/story")
	var parseError *ParseError
	if r != nil || !errors.Is(err, ErrDefinitionTransformLimit) || !errors.As(err, &parseError) || parseError.Code != ErrExtract {
		t.Fatalf("JSON depth limit did not surface: result=%v error=%v", r, err)
	}

	optionalDepth := transformDefinitions(t, `    - target: descendants
      selector: .media
      attribute.set_from:
        name: src
        value:
          json:
            attribute: data-props
            path: [{field: a}]
`)
	r, err = New(WithDefinitions(optionalDepth)).ParseHTML(context.Background(), `<article><p>Optional bounded JSON depth.</p><div class="media" data-props='`+deep+`'>Media</div></article>`, "https://93.184.216.34/story")
	if r != nil || !errors.Is(err, ErrDefinitionTransformLimit) || !errors.As(err, &parseError) || parseError.Code != ErrExtract {
		t.Fatalf("optional JSON depth limit was swallowed: result=%v error=%v", r, err)
	}

	oversized := transformDefinitions(t, `    - target: descendants
      selector: .media
      attribute.set_from:
        name: src
        value:
          json:
            attribute: data-props
            path: [{field: src}]
            required: true
`)
	large := `{"src":"` + strings.Repeat("a", DefinitionCapabilities().MaxValueBytes) + `"}`
	r, err = New(WithDefinitions(oversized)).ParseHTML(context.Background(), `<article><p>Bounded JSON bytes.</p><div class="media" data-props='`+large+`'>Media</div></article>`, "https://93.184.216.34/story")
	if r != nil || !errors.Is(err, ErrDefinitionTransformLimit) {
		t.Fatalf("JSON byte limit did not surface: result=%v error=%v", r, err)
	}
}

func TestDefinitionsNamedAbendblattAlgorithm(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: p
      algorithm.apply: {name: abendblatt.deobfuscate}
`)
	source := `<article><p class="obfuscated">Ifmmp² &lt;tdsjqu&gt;</p><p>Unchanged article paragraph.</p></article>`
	for _, format := range []string{"html", "markdown", "text"} {
		r, err := New(WithDefinitions(s), WithContentType(format)).ParseHTML(context.Background(), source, "https://93.184.216.34/story")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.Content, "Hello! ;script=") || !strings.Contains(r.Content, "Unchanged article paragraph.") {
			t.Fatalf("%s did not decode expected text: %s", format, r.Content)
		}
		if strings.Contains(r.Content, "<script>") {
			t.Fatalf("%s decoded text became active markup: %s", format, r.Content)
		}
	}
}
