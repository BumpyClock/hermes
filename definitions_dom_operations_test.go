package hermes

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefinitionsDOMRetainMoveAndCreate(t *testing.T) {
	s := transformDefinitions(t, `    - target: root
      element.retain:
        selector: ".keep, .destination"
        select: all
        required: true
    - target: root
      element.move:
        source: {selector: img, select: all}
        target: {selector: .destination, select: first}
        position: append
        required: true
    - target: root
      element.create:
        target: {selector: .destination, select: all}
        position: prepend
        node:
          tag: p
          text: "Constructed literal <not-markup>"
          attributes: {title: "Safe literal"}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><p>Discarded article text.</p><figure class="keep"><img src="/photo.jpg" alt="Photo"><figcaption>Caption retained.</figcaption></figure><div class="destination"></div></article>`, "https://93.184.216.34/story")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Caption retained.", "Constructed literal &lt;not-markup&gt;", `src="https://93.184.216.34/photo.jpg"`} {
		if !strings.Contains(r.Content, want) {
			t.Errorf("missing %q: %s", want, r.Content)
		}
	}
	if strings.Contains(r.Content, "Discarded article text") || strings.Count(r.Content, "photo.jpg") != 1 {
		t.Fatalf("retain/move ordering lost: %s", r.Content)
	}
}

func TestDefinitionsDOMSelectionModesAndDuplicateReferences(t *testing.T) {
	s := transformDefinitions(t, `    - target: root
      element.move:
        source: {selector: "figure, figure img", select: all}
        target: {selector: .slot, select: all}
        position: append
        required: true
    - target: root
      element.create:
        target: {selector: .slot, select: last}
        position: append
        node: {tag: p, text: "Last slot marker"}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><figure><img src="/only.jpg" alt="Only"><figcaption>Only caption</figcaption></figure><div class="slot"></div><div class="slot"></div></article>`, "https://93.184.216.34/story")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(r.Content, "Only caption") != 2 || strings.Count(r.Content, "only.jpg") != 2 || strings.Count(r.Content, "Last slot marker") != 1 {
		t.Fatalf("all/last selection or duplicate source handling failed: %s", r.Content)
	}
}

func TestDefinitionsNoscriptRecovery(t *testing.T) {
	s := transformDefinitions(t, `    - target: root
      noscript.recover:
        source: {selector: noscript, select: all}
        target: {selector: .slot, select: first}
        position: replace
        required: true
`)
	for _, source := range []string{
		`<article><p>Parsed media report.</p><div class="slot"></div><noscript><picture><img src="/parsed.jpg" alt="Parsed image"></picture></noscript></article>`,
		`<article><p>Encoded media report.</p><div class="slot"></div><noscript>&lt;picture&gt;&lt;img src="/encoded.jpg" alt="Encoded image"&gt;&lt;/picture&gt;</noscript></article>`,
		`<article><p>Malformed media report.</p><div class="slot"></div><noscript>&lt;picture&gt;&lt;img src="/malformed.jpg" alt="Malformed image"&gt;</noscript></article>`,
	} {
		r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), source, "https://93.184.216.34/story")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.Content, `<img`) || strings.Contains(r.Content, "<noscript") {
			t.Fatalf("noscript image was not safely recovered: %s", r.Content)
		}
	}
}

func TestDefinitionsDOMOperationsRejectCyclesAndUnsafeConstruction(t *testing.T) {
	cycle := transformDefinitions(t, `    - target: root
      element.move:
        source: {selector: .box, select: first}
        target: {selector: .inside, select: first}
        position: append
        required: true
`)
	r, err := New(WithDefinitions(cycle)).ParseHTML(context.Background(), `<article><div class="box"><div class="inside">Nested target</div></div></article>`, "https://93.184.216.34/story")
	var pe *ParseError
	var de *DefinitionOperationError
	if r != nil || !errors.As(err, &pe) || pe.Code != ErrExtract || !errors.As(err, &de) || de.Step != 1 {
		t.Fatalf("cycle did not return contextual extraction error: result=%v error=%v", r, err)
	}

	dir := t.TempDir()
	if writeErr := os.WriteFile(filepath.Join(dir, "invalid.yaml"), []byte(`schema: 1
site: invalid-construction
hosts: [93.184.216.34]
content:
  groups: [[article]]
  transforms:
    - target: root
      element.create:
        target: {selector: p}
        position: append
        node: {tag: script, text: "alert(1)"}
`), 0o600); writeErr != nil {
		t.Fatal(writeErr)
	}
	_, err = LoadDefinitions(dir)
	var diagnostic *DefinitionError
	if !errors.As(err, &diagnostic) || !strings.Contains(diagnostic.Field, "element.create.node.tag") {
		t.Fatalf("unsafe construction accepted: %v", err)
	}
}

func TestDefinitionsDOMOptionalDetachedTargetAndAncestorMove(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: .discarded-slot
      element.remove: {}
    - target: root
      element.move:
        source: {selector: .optional-photo, select: first}
        target: {selector: .discarded-slot, select: first}
        position: append
    - target: root
      element.move:
        source: {selector: .kept-photo, select: first}
        target: {selector: figure, select: first}
        position: prepend
        required: true
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><div class="discarded-slot"></div><img class="optional-photo" src="/optional.jpg" alt="Optional"><figure><figcaption>Caption after moved image.</figcaption><img class="kept-photo" src="/kept.jpg" alt="Kept"></figure></article>`, "https://93.184.216.34/story")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(r.Content, "optional.jpg") != 1 || strings.Count(r.Content, "kept.jpg") != 1 || strings.Index(r.Content, "kept.jpg") > strings.Index(r.Content, "Caption after") {
		t.Fatalf("detached optional target or ancestor move was not stable: %s", r.Content)
	}
}

func TestDefinitionsDOMGeneratedNodeLimit(t *testing.T) {
	s := transformDefinitions(t, `    - target: root
      element.create:
        target: {selector: .slot, select: all}
        position: append
        node: {tag: p, text: "Generated"}
`)
	source := "<article>" + strings.Repeat(`<div class="slot"></div>`, DefinitionCapabilities().MaxContentNodes/2) + "</article>"
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), source, "https://93.184.216.34/story")
	var pe *ParseError
	if r != nil || !errors.Is(err, ErrDefinitionTransformLimit) || !errors.As(err, &pe) || pe.Code != ErrExtract {
		t.Fatalf("generated node limit was not surfaced: result=%v error=%v", r, err)
	}
}

func TestDefinitionsCleanerPreserveDoesNotBypassSanitization(t *testing.T) {
	s, _ := localDefinitions(t, `schema: 1
site: preserved-media
hosts: [93.184.216.34]
content:
  groups: [[article]]
  preserve: [.media, script]
  default_cleaner: true
`)
	if rule := s.snapshot.Match("93.184.216.34"); rule == nil || len(rule.Content.Preserve) != 2 {
		t.Fatalf("preserve selectors were not loaded: %#v", rule)
	}
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><p>Article context remains visible.</p><div class="media"><img src="/photo.jpg" width="1" height="1" alt="Article image"></div><script>alert(1)</script></article>`, "https://93.184.216.34/story")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Content, "photo.jpg") || strings.Contains(r.Content, "script") || strings.Contains(r.Content, "alert(1)") {
		t.Fatalf("preservation weakened sanitation or lost media: %s", r.Content)
	}
}
