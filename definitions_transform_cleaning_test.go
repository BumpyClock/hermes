package hermes

import (
	"context"
	"strings"
	"testing"
)

func TestDefinitionsTransformCleaningOrder(t *testing.T) {
	for _, mode := range []string{"true", "false"} {
		s, _ := localDefinitions(t, `schema: 1
site: cleaning-order
hosts: [93.184.216.34]
metadata:
  title: [{text: h1}]
content:
  groups: [[article]]
  default_cleaner: `+mode+`
  transforms:
    - target: descendants
      selector: aside
      element.rename: {tag: h1}
    - target: descendants
      selector: "a[href='/ads/promo']"
      string.replace: {attribute: href, old: /ads/, new: /remove/}
  remove: ["a[href='/remove/promo']"]
`)
		r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><h1>Main report</h1><p>The primary article contains substantial reporting to exercise the actual default cleaner and preserve the principal story.</p><aside>Optional sidebar transformed before cleaning.</aside><a href="/ads/promo">Explicitly removed after transformation.</a><p>A second substantive paragraph supplies further context and details that the cleanup process must preserve.</p></article>`, "https://93.184.216.34/story")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(r.Content, "Explicitly removed") || !strings.Contains(r.Content, "primary article") {
			t.Fatalf("site-removal ordering: %s", r.Content)
		}
		if strings.Contains(r.Content, "Optional sidebar") != (mode == "false") {
			t.Fatalf("default cleaner %s did not observe transformed tag: %s", mode, r.Content)
		}
	}
}

func TestDefinitionsTransformRetiredRootsAreNotRematched(t *testing.T) {
	s := transformDefinitions(t, `    - target: root
      element.unwrap: {}
    - target: root
      attribute.copy: {from: missing, to: title, required: true}
    - target: descendants
      selector: p
      element.rename: {tag: blockquote}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), `<article><p>Promoted children remain descendants.</p></article>`, "https://93.184.216.34/story")
	if err != nil || !strings.Contains(r.Content, "<blockquote>Promoted children") {
		t.Fatalf("retired root or promoted children misclassified: %+v, %v", r, err)
	}
}
