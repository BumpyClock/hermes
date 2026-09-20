package hermes

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDefinitionsNYTimesOrderedImageRepair(t *testing.T) {
	s, _ := localDefinitions(t, `schema: 1
site: nytimes-synthetic
hosts: [93.184.216.34]
metadata:
  title: [{text: h1}]
  lead_image_url:
    - attribute: {selector: "img[src='../photo-{{size}}.jpg']", name: src}
content:
  groups: [[article]]
  default_cleaner: false
  transforms:
    - target: descendants
      selector: img.g-lazy
      string.replace: {attribute: src, old: '{{size}}', new: '640'}
    - target: descendants
      selector: "img[src='../photo-640.jpg']"
      attribute.set: {name: alt, value: Repaired photograph}
    - target: descendants
      selector: aside
      element.rename: {tag: p}
  remove: ["a[href='/ads/offer']"]
`)
	source := `<html><head><base target="_blank"><base href="/assets/images/"><base href="/wrong/"></head><body><article>
<h1>Synthetic NYTimes report</h1><p>An independently written report supplies meaningful article content.</p>
<img class="g-lazy" src="../photo-{{size}}.jpg" width="640" height="480">
<aside>Caption transformed by the shared executor.</aside><a href="/ads/offer">Excluded advert</a>
<p>Literal &lt;script&gt;example&lt;/script&gt; and &amp;copy; remain text.</p></article></body></html>`
	for _, format := range []string{"html", "markdown", "text"} {
		r, err := New(WithDefinitions(s), WithContentType(format)).ParseHTML(context.Background(), source, "https://93.184.216.34/news/story")
		if err != nil {
			t.Fatal(err)
		}
		if r.Title != "Synthetic NYTimes report" || !strings.Contains(r.LeadImageURL, "photo-") || strings.Contains(r.LeadImageURL, "photo-640") {
			t.Fatalf("source metadata mutated: %+v", r)
		}
		if !strings.Contains(r.Content, "independently written") || strings.Contains(r.Content, "Excluded advert") {
			t.Fatalf("%s article boundary: %s", format, r.Content)
		}
		if format != "text" && (!strings.Contains(r.Content, "https://93.184.216.34/assets/photo-640.jpg") || !strings.Contains(r.Content, "Repaired photograph")) {
			t.Fatalf("%s ordered repair/base: %s", format, r.Content)
		}
		if format == "html" && (!strings.Contains(r.Content, "<p>Caption transformed") || strings.Contains(r.Content, "<script>example")) {
			t.Fatalf("rename/sanitization: %s", r.Content)
		}
		if format == "markdown" && (strings.Contains(r.Content, "<script>") || !strings.Contains(r.Content, "&amp;copy;")) {
			t.Fatalf("literal Markdown safety: %s", r.Content)
		}
	}
}

func transformDefinitions(t *testing.T, steps string) *Definitions {
	t.Helper()
	s, _ := localDefinitions(t, `schema: 1
site: transforms
hosts: [93.184.216.34]
content:
  groups: [[article, .second-root]]
  default_cleaner: false
  transforms:
`+steps)
	return s
}

func TestDefinitionsTransformFamilies(t *testing.T) {
	tests := []struct{ name, steps, source, want string }{
		{"attribute copy then remove", `    - target: descendants
      selector: img
      attribute.copy: {from: data-photo, to: src, required: true}
    - target: descendants
      selector: img
      attribute.remove: {name: data-photo}
`, `<img data-photo="/photo.jpg" alt="Copied image">`, `img[src='https://93.184.216.34/photo.jpg']:not([data-photo])`},
		{"attribute empty set", `    - target: descendants
      selector: img
      attribute.set: {name: alt, value: ''}
`, `<img src="/photo.jpg" alt="old">`, `img[alt='']`},
		{"regex and URL construction", `    - target: descendants
      selector: a
      regex.capture: {attribute: data-id, pattern: '^photo:([a-z/ ]+)$', group: 1, required: true}
    - target: descendants
      selector: a
      url.build:
        attribute: href
        base: https://media.example.com/items?fixed=yes
        path: [{literal: image}, {attribute: data-id, required: true}]
        query:
          - name: q
            value: {attribute: title, required: true}
          - name: size
            value: {literal: '640'}
`, `<a data-id="photo:a/b c" title="a&amp;b">Media</a>`, `a[href='https://media.example.com/items/image/a%2Fb%20c?fixed=yes&q=a%26b&size=640']`},
		{"URL resolution", `    - target: descendants
      selector: a
      url.resolve: {attribute: href, required: true}
    - target: descendants
      selector: "a[href='https://93.184.216.34/photo']"
      element.rename: {tag: strong}
`, `<a href="../photo">Link</a>`, `strong`},
		{"root rename and descendant exclusion", `    - target: descendants
      selector: article
      element.remove: {}
    - target: root
      element.rename: {tag: blockquote}
`, `<p>A surviving child</p>`, `blockquote > p`},
		{"nested unwrap", `    - target: descendants
      selector: div
      element.unwrap: {}
`, `<div><div><p>First</p></div><p>Second</p></div><div><p>Third</p></div>`, `body > p:nth-child(3)`},
		{"root unwrap", `    - target: root
      element.unwrap: {}
`, `<p>First</p><p>Second</p>`, `body > p:nth-child(2)`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := transformDefinitions(t, test.steps)
			r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), "<article>"+test.source+"</article>", "https://93.184.216.34/news/story")
			if err != nil {
				t.Fatal(err)
			}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(r.Content))
			if err != nil {
				t.Fatal(err)
			}
			if doc.Find(test.want).Length() != 1 {
				t.Fatalf("missing %s: %s", test.want, r.Content)
			}
			if strings.Contains(test.name, "unwrap") && strings.Index(r.Content, "First") > strings.Index(r.Content, "Second") {
				t.Fatalf("sibling order lost: %s", r.Content)
			}
		})
	}
}

func TestDefinitionsTransformConditions(t *testing.T) {
	tests := []struct {
		name, condition, source string
		match                   bool
	}{
		{"existence", "exists: {attribute: title}", `<p title="">Hello</p>`, true},
		{"missing existence", "exists: {attribute: title}", `<p>Hello</p>`, false},
		{"negated existence", "exists: {attribute: title, present: false}", `<p>Hello</p>`, true},
		{"text content", "text: {comparison: equals, value: Hello world}", `<p>Hello <b>world</b></p>`, true},
		{"literal text", "text: {comparison: contains, value: '<script>'}", `<p>&lt;script&gt;example</p>`, true},
		{"text prefix", "text: {comparison: prefix, value: Hell}", `<p>Hello</p>`, true},
		{"text suffix attribute", "text: {attribute: title, comparison: suffix, value: world}", `<p title="Hello world">Hello</p>`, true},
		{"text unmet", "text: {comparison: equals, value: Hello}", `<p>Different</p>`, false},
		{"text optional", "text: {attribute: title, comparison: equals, value: Hello}", `<p>Hello</p>`, false},
		{"number gt", "number: {attribute: data-count, comparison: gt, value: 4}", `<p data-count="5">Hello</p>`, true},
		{"number ge", "number: {attribute: data-count, comparison: ge, value: 5}", `<p data-count="5">Hello</p>`, true},
		{"number lt", "number: {attribute: data-count, comparison: lt, value: 6.1}", `<p data-count=" 5.5 ">Hello</p>`, true},
		{"number le", "number: {attribute: data-count, comparison: le, value: 5}", `<p data-count="5">Hello</p>`, true},
		{"number equal", "number: {attribute: data-count, comparison: eq, value: 5}", `<p data-count="5">Hello</p>`, true},
		{"number unequal", "number: {attribute: data-count, comparison: ne, value: 5}", `<p data-count="6">Hello</p>`, true},
		{"number unmet", "number: {attribute: data-count, comparison: gt, value: 5}", `<p data-count="4">Hello</p>`, false},
		{"number optional", "number: {attribute: data-count, comparison: eq, value: 0}", `<p>Hello</p>`, false},
		{"descendant", "descendant: {selector: b}", `<p>Hello <b>world</b></p>`, true},
		{"descendant excludes self", "descendant: {selector: p}", `<p>Hello</p>`, false},
		{"descendant absent", "descendant: {selector: img, present: false}", `<p>Hello</p>`, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := transformDefinitions(t, "    - target: descendants\n      selector: p\n      when:\n        - "+test.condition+"\n      element.rename: {tag: blockquote}\n")
			r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), "<article>"+test.source+"</article>", "https://93.184.216.34/story")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(r.Content, "<blockquote") != test.match {
				t.Fatalf("predicate result: %s", r.Content)
			}
		})
	}
}

func TestDefinitionsTransformSnapshotOrderingAndIsolation(t *testing.T) {
	s := transformDefinitions(t, `    - target: descendants
      selector: div
      when: [{descendant: {selector: div, present: false}}]
      element.unwrap: {}
    - target: root
      when: [{descendant: {selector: div, present: false}}]
      attribute.set: {name: data-ready, value: yes}
    - target: descendants
      selector: "[data-ready] p"
      attribute.set: {name: title, value: first}
    - target: descendants
      selector: "p[title='first']"
      attribute.set: {name: title, value: second}
    - target: descendants
      selector: "p[title='second']"
      when:
        - text: {comparison: prefix, value: Keep}
        - exists: {attribute: title}
      element.rename: {tag: blockquote}
`)
	source := `<article><div><div><p>Keep first.</p></div><p>Keep second.</p></div><div><p>Keep third.</p></div></article><section class="second-root"><p>Keep fourth.</p></section>`
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), source, "https://93.184.216.34/story")
			if err != nil {
				t.Error(err)
				return
			}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(r.Content))
			if err != nil {
				t.Error(err)
				return
			}
			if doc.Find("blockquote").Length() != 4 || doc.Find("p").Length() != 0 {
				t.Errorf("not deepest-first or later steps lost: %s", r.Content)
			}
			prior := -1
			for _, value := range []string{"Keep first.", "Keep second.", "Keep third.", "Keep fourth."} {
				index := strings.Index(r.Content, value)
				if index <= prior {
					t.Errorf("document-order tie changed: %s", r.Content)
				}
				prior = index
			}
		}()
	}
	wg.Wait()
}

func TestDefinitionsTransformRemovalPreservesFallbackSource(t *testing.T) {
	s := transformDefinitions(t, "    - target: root\n      element.remove: {}\n")
	source := `<article><p>Source fallback still contains a substantive paragraph with enough reporting to survive extraction.</p><p>Further source information remains available after the selected content copy was removed.</p></article>`
	r, err := New(WithDefinitions(s)).ParseHTML(context.Background(), source, "https://93.184.216.34/story")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Content, "Source fallback") || !strings.Contains(r.Content, "Further source") {
		t.Fatalf("fallback source mutated: %s", r.Content)
	}
}
