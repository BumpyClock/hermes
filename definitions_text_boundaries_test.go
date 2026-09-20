package hermes

import (
	"context"
	"strings"
	"testing"
)

func TestDefinitionTextPreservesBlockBoundaries(t *testing.T) {
	snapshot, _ := localDefinitions(t, `schema: 1
site: text-boundaries
hosts: [93.184.216.34]
content:
  groups: [[article]]
  default_cleaner: false
`)
	const source = `<html><head><title>Original boundary report</title></head><body><article>` +
		`<p>First paragraph</p><p>Second<br>line</p><ul><li>First item</li><li>Second item</li></ul>` +
		`<table><caption>Measurements</caption><tr><td>First cell</td><td>Second cell</td></tr></table>` +
		`<p>pre<strong>fix</strong> &lt;literal&gt;</p></article></body></html>`
	for _, format := range []string{"text", "txt", "text/plain"} {
		result, err := New(WithDefinitions(snapshot), WithContentType(format)).ParseHTML(
			context.Background(), source, "https://93.184.216.34/article")
		if err != nil {
			t.Fatal(err)
		}
		for _, expected := range []string{
			"First paragraph Second line", "First item Second item",
			"Measurements First cell Second cell", "prefix <literal>",
		} {
			if !strings.Contains(result.Content, expected) {
				t.Fatalf("%s: missing boundary-preserving text %q in %q", format, expected, result.Content)
			}
		}
	}
}
