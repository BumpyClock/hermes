package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"golang.org/x/net/html"
)

func TestMarkdownWrappersPreserveDocumentEdgeUnicodeSpacing(t *testing.T) {
	for _, tag := range []string{"em", "i", "strong", "b"} {
		for _, space := range []string{"\u00a0", "\u202f", "\u3000"} {
			for _, edges := range [][2]string{{space, ""}, {"", space}, {space, space}} {
				source := "<p><" + tag + ">" + edges[0] + "Read" + edges[1] + "</" + tag + "></p>"
				markdown := convertToMarkdown(source)
				var rendered bytes.Buffer
				if err := goldmark.Convert([]byte(markdown), &rendered); err != nil {
					t.Fatal(err)
				}
				document, err := html.Parse(&rendered)
				if err != nil {
					t.Fatal(err)
				}
				want := edges[0] + "Read" + edges[1]
				if got := strings.Trim(documentText(document), "\n"); got != want {
					t.Fatalf("source %q produced visible text %q, want %q (Markdown %q)", source, got, want, markdown)
				}
				semanticTag := "em"
				if tag == "strong" || tag == "b" {
					semanticTag = "strong"
				}
				if !documentContainsTag(document, semanticTag) {
					t.Fatalf("source %q lost %s semantics in %q", source, semanticTag, markdown)
				}
			}
		}
	}
}
