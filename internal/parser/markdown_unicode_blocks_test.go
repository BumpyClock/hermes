package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"golang.org/x/net/html"
)

func TestMarkdownBlocksPreserveUnicodeSpacing(t *testing.T) {
	for _, space := range []string{"\u00a0", "\u202f", "\u3000"} {
		for _, source := range []string{
			"<figure><figcaption>" + space + "Caption" + space + "</figcaption></figure>",
			"<table><caption>" + space + "Caption" + space + "</caption><tr><td>Cell</td></tr></table>",
			"<table><tr><th>" + space + "Cell" + space + "</th></tr></table>",
			"<table><tr><td>" + space + "Cell" + space + "</td></tr></table>",
		} {
			markdown := convertToMarkdown(source)
			var rendered bytes.Buffer
			if err := goldmark.Convert([]byte(markdown), &rendered); err != nil {
				t.Fatal(err)
			}

			document, err := html.Parse(&rendered)
			if err != nil {
				t.Fatal(err)
			}
			want := space + "Cell" + space
			if strings.Contains(source, "Caption") {
				want = space + "Caption" + space
			}
			if got := documentText(document); !strings.Contains(got, want) {
				t.Fatalf("source %q produced visible text %q without %q (Markdown %q)", source, got, want, markdown)
			}
		}
	}
}
