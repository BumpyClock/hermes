package hermes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestDefinitionsTransformPublicFailures(t *testing.T) {
	tests := []struct{ name, operation, source, cause string }{
		{"required absent", "attribute.copy: {from: data-src, to: src, required: true}", `<img src="/safe.jpg">`, "required attribute"},
		{"required empty", "string.replace: {attribute: src, old: x, new: y, required: true}", `<img src=" ">`, "empty"},
		{"invalid URL", "url.resolve: {attribute: src}", `<img src="%zz">`, "invalid URL escape"},
		{"unsafe URL", "url.resolve: {attribute: src}", `<img src="javascript:alert(1)">`, "HTTP(S)"},
		{"required capture", "regex.capture: {attribute: src, pattern: '^image-(.+)$', group: 1, required: true}", `<img src="missing">`, "did not match"},
		{"bad number", "when: [{number: {attribute: width, comparison: gt, value: 0}}]\n      attribute.set: {name: alt, value: Changed}", `<img src="/safe.jpg" width="invalid">`, "invalid syntax"},
		{"nonfinite number", "when: [{number: {attribute: width, comparison: gt, value: 0}}]\n      attribute.set: {name: alt, value: Changed}", `<img src="/safe.jpg" width="NaN">`, "finite"},
		{"runtime path", "url.build: {attribute: src, base: 'https://media.example.com', path: [{attribute: title, required: true}]}", `<img title="..">`, "path segment"},
	}
	const articleURL = "https://93.184.216.34/story"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := transformDefinitions(t, "    - target: descendants\n      selector: img\n      "+test.operation+"\n")
			source := "<article><p>Enough article text must not hide a genuine operation failure behind generic fallback.</p>" + test.source + "</article>"
			c := New(WithDefinitions(s), WithTransport(definitionTransport(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(source)), Request: r}, nil
			})))
			for _, op := range []string{"Parse", "ParseHTML"} {
				var r *Result
				var err error
				if op == "Parse" {
					r, err = c.Parse(context.Background(), articleURL)
				} else {
					r, err = c.ParseHTML(context.Background(), source, articleURL)
				}
				var pe *ParseError
				var de *DefinitionOperationError
				if r != nil || !errors.As(err, &pe) || pe.Code != ErrExtract || pe.Op != op || pe.URL != articleURL {
					t.Fatalf("%s lost public classification: result=%v error=%v", op, r, err)
				}
				if !errors.As(err, &de) || de.Site != "transforms" || de.Step != 1 || de.Source == "" || de.Line == 0 || de.Column == 0 || errors.Unwrap(de) == nil || !strings.Contains(err.Error(), test.cause) {
					t.Fatalf("lost definition context/cause: %v", err)
				}
				if test.name == "invalid URL" {
					var cause *url.Error
					if !errors.As(err, &cause) {
						t.Fatal("URL cause lost")
					}
				}
				if test.name == "bad number" {
					var cause *strconv.NumError
					if !errors.As(err, &cause) {
						t.Fatal("numeric cause lost")
					}
				}
			}
		})
	}
}

func TestDefinitionsTransformOptionalMisses(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: .not-present
      attribute.copy: {from: missing, to: title, required: true}
    - target: descendants
      selector: p
      attribute.copy: {from: missing, to: title}
    - target: descendants
      selector: p
      regex.capture: {attribute: title, pattern: '(not-present)', group: 1}
    - target: descendants
      selector: p
      url.build: {attribute: title, base: 'https://example.com', path: [{attribute: missing}]}
    - target: descendants
      selector: p
      when: [{exists: {attribute: missing}}]
      url.resolve: {attribute: missing, required: true}
    - target: descendants
      selector: "p[title='Original']"
      element.rename: {tag: blockquote}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><p title="Original">Optional absence and unmet predicates leave this article intact.</p></article>`, "https://93.184.216.34/story")
	if err != nil || !strings.Contains(r.Content, `<blockquote>Optional absence`) {
		t.Fatalf("normal absence failed: %+v, %v", r, err)
	}
}

func TestDefinitionsTransformRuntimeLimits(t *testing.T) {
	support := DefinitionCapabilities()
	tests := []struct{ name, source, operation string }{
		{"oversized input", `<p title="` + strings.Repeat("a", support.MaxValueBytes+1) + `">Value</p>`, "string.replace: {attribute: title, old: a, new: b}"},
		{"oversized output", `<p title="` + strings.Repeat("a", support.MaxValueBytes/2+1) + `">Value</p>`, "string.replace: {attribute: title, old: a, new: bb}"},
		{"node count", strings.Repeat("<p>Value</p>", support.MaxContentNodes/2+1), "attribute.set: {name: title, value: Changed}"},
		{"depth", strings.Repeat("<div>", support.MaxContentDepth+1) + "<p>Value</p>" + strings.Repeat("</div>", support.MaxContentDepth+1), "attribute.set: {name: title, value: Changed}"},
		{"work budget", strings.Repeat(`<p title="`+strings.Repeat("a", 40000)+`">Value</p>`, support.MaxTransformWork/40000+2), "string.replace: {attribute: title, old: a, new: b}"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			steps := "    - target: descendants\n      selector: p\n      " + test.operation + "\n"
			if test.name == "work budget" {
				steps += steps
			}
			s := transformDefinitions(t, steps)
			r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), "<article>"+test.source+"</article>", "https://93.184.216.34/story")
			var pe *ParseError
			if r != nil || !errors.Is(err, ErrDefinitionTransformLimit) || !errors.As(err, &pe) || pe.Code != ErrExtract {
				t.Fatalf("limit not surfaced: %+v, %v", r, err)
			}
		})
	}
}

func TestDefinitionsTransformsCannotBypassSanitization(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: a
      attribute.set: {name: href, value: 'javascript:alert(1)'}
    - target: descendants
      selector: p
      attribute.set: {name: onclick, value: 'alert(1)'}
    - target: descendants
      selector: p
      element.rename: {tag: h2}
`)
	for _, format := range []string{"html", "markdown", "text"} {
		r, err := New(WithDefinitions(s), WithContentType(format)).ParseHTML(context.Background(), `<article><p>Safe &lt;script&gt;example&lt;/script&gt; &amp;copy;</p><a href="/safe">Link text</a></article>`, "https://93.184.216.34/story")
		if err != nil {
			t.Fatal(err)
		}
		for _, unsafe := range []string{"javascript:", "onclick", "alert(1)"} {
			if strings.Contains(r.Content, unsafe) {
				t.Errorf("%s leaked %s: %s", format, unsafe, r.Content)
			}
		}
		if format == "markdown" && !strings.Contains(r.Content, "## Safe &lt;script&gt;example&lt;/script&gt; &amp;copy;") {
			t.Fatalf("Markdown formatting or literal text safety lost: %s", r.Content)
		}
	}
}

func TestDefinitionTransformCapabilitiesAreIndependent(t *testing.T) {
	s := DefinitionCapabilities()
	for _, capability := range []string{"metadata.text_capture", "transform.element.rename", "transform.element.retain", "transform.element.move", "transform.element.create", "transform.noscript.recover", "transform.attribute.set_from", "transform.algorithm.apply", "transform.url.resolve.base", "transform.url.resolve.to", "element.create.typed_attributes", "element.create.self_target", "transform.noscript.recover.self", "value.descendant_attribute", "value.json", "content.preserve", "transform.url.build", "condition.number"} {
		if !strings.Contains(fmt.Sprint(s.Capabilities), capability) {
			t.Errorf("missing %s", capability)
		}
	}
	if len(s.Algorithms) != 1 || s.Algorithms[0] != "abendblatt.deobfuscate" {
		t.Fatalf("unexpected algorithms: %v", s.Algorithms)
	}
	s.Capabilities[len(s.Capabilities)-1] = "mutated"
	s.Algorithms[0] = "mutated"
	fresh := DefinitionCapabilities()
	if strings.Contains(fmt.Sprint(fresh.Capabilities), "mutated") || strings.Contains(fmt.Sprint(fresh.Algorithms), "mutated") {
		t.Fatal("support storage escaped")
	}
}
