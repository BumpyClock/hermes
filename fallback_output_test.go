package hermes

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDefinitionsLastResortTextFormats(t *testing.T) {
	unmatched, _ := localDefinitions(t, `schema: 1
site: unmatched
hosts: [example.com]
content:
  groups: [[article]]
  default_cleaner: false
`)
	for name, snapshot := range map[string]*Definitions{"nil": nil, "unmatched": unmatched} {
		for _, tc := range []struct {
			name string
			text string
			html string
			md   string
		}{
			{"raw element", "<img src=x onerror=alert(1)>", "&lt;img src=x onerror=alert(1)&gt;", `\<img src\=x onerror\=alert\(1\)\>`},
			{"literal punctuation", "Compare <angle> & values > 2", "Compare &lt;angle&gt; &amp; values &gt; 2", `Compare \<angle\> \& values \> 2`},
			{"meaningful text", "A meaningful fallback report preserves readable article content.", "A meaningful fallback report preserves readable article content.", `A meaningful fallback report preserves readable article content\.`},
			{"literal entities", "Keep &amp; and &copy; literal", "Keep &amp;amp; and &amp;copy; literal", `Keep \&amp\; and \&copy\; literal`},
			{"markdown syntax", "[label](javascript:alert(1)) *literal* `code`", "[label](javascript:alert(1)) *literal* `code`", `\[label\]\(javascript\:alert\(1\)\) \*literal\* ` + "\\`code\\`"},
		} {
			for format, want := range map[string]string{"html": tc.html, "markdown": tc.md, "text": tc.text} {
				t.Run(name+"/"+tc.name+"/"+format, func(t *testing.T) {
					source := "<title>Review article</title><main><iframe>" + tc.text + "</iframe></main>"
					result, err := New(WithDefinitions(snapshot), WithContentType(format)).ParseHTML(context.Background(), source, "https://93.184.216.34/review")
					if err != nil {
						t.Fatal(err)
					}
					if result.Content != want {
						t.Errorf("content = %q, want %q", result.Content, want)
					}
					if format == "html" {
						doc, err := goquery.NewDocumentFromReader(strings.NewReader(result.Content))
						if err != nil {
							t.Fatal(err)
						}
						if doc.Find("body").Text() != tc.text || doc.Find("body *").Length() != 0 {
							t.Errorf("fallback must remain literal text, got %q", result.Content)
						}
					}
				})
			}
		}
	}
}
