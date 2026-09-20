package hermes

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestDefinitionsTransformsPreserveNormalFullArticleStructure(t *testing.T) {
	s := transformDefinitions(t, `    - target: root
      element.rename: {tag: article}
`)
	source := calibratedFullArticleSource()
	document, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	nodes, initialWork := 0, 0
	var count func(*html.Node, int)
	count = func(node *html.Node, _ int) {
		nodes++
		initialWork += 1 + len(node.Data)
		for _, attribute := range node.Attr {
			initialWork += len(attribute.Key) + len(attribute.Val) + 1
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			count(child, 0)
		}
	}
	count(document, 0)
	support := DefinitionCapabilities()
	if nodes <= 10000 || nodes >= support.MaxContentNodes || initialWork < 750000 || initialWork >= support.MaxTransformWork {
		t.Fatalf("synthetic article no longer calibrates the intended range: nodes=%d work=%d support=%+v", nodes, initialWork, support)
	}
	t.Logf("calibrated synthetic article: nodes=%d initial_work=%d", nodes, initialWork)

	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), source, "https://93.184.216.34/story")
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"Heading 0", "Second 550", "Caption 1099.", "Value 1099"} {
		if !strings.Contains(r.Content, marker) {
			t.Fatalf("full article structure lost %q", marker)
		}
	}
	for _, tag := range []string{"<h2>", "<ul>", "<figure>", "<table>"} {
		if !strings.Contains(r.Content, tag) {
			t.Fatalf("full article semantic structure lost %s", tag)
		}
	}
}

func TestDefinitionsShapePreservingStepsAvoidFullTreeRescans(t *testing.T) {
	s := transformDefinitions(t, `    - target: root
      element.rename: {tag: article}
    - target: root
      element.rename: {tag: article}
    - target: descendants
      selector: h2
      attribute.set: {name: title, value: "article heading"}
    - target: descendants
      selector: h2
      attribute.copy: {from: title, to: data-heading, required: true}
    - target: descendants
      selector: h2
      string.replace: {attribute: data-heading, old: " ", new: "-"}
`)
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), calibratedFullArticleSource(), "https://93.184.216.34/story")
	if err != nil || !strings.Contains(r.Content, "Heading 1099") {
		t.Fatalf("shape-preserving transforms exhausted the work budget: result=%v error=%v", r, err)
	}
}

func calibratedFullArticleSource() string {
	const sections = 1100
	layout := strings.Repeat("layout-", 90)
	var source strings.Builder
	source.WriteString("<article>")
	for i := 0; i < sections; i++ {
		index := strconv.Itoa(i)
		source.WriteString(`<section data-layout="` + layout + `"><h2>Heading ` + index + `</h2><p>Paragraph ` + index + `.</p><ul><li>First ` + index + `</li><li>Second ` + index + `</li></ul><figure><img src="/image-` + index + `.jpg" alt="Image ` + index + `"><figcaption>Caption ` + index + `.</figcaption></figure><table><tbody><tr><th>Key ` + index + `</th><td>Value ` + index + `</td></tr></tbody></table></section>`)
	}
	source.WriteString("</article>")
	return source.String()
}
