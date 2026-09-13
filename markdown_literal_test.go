package hermes

import (
	"context"
	"strings"
	"testing"
)

func TestDefinitionsMarkdownLiteralText(t *testing.T) {
	snapshots := map[string]*Definitions{"nil": nil, "legacy-generic": nil}
	for _, cleaner := range []string{"true", "false"} {
		snapshot, _ := localDefinitions(t, `schema: 1
site: literal
hosts: [93.184.216.34]
content:
  groups: [[article]]
  default_cleaner: `+cleaner+"\n")
		snapshots["selected-"+cleaner] = snapshot
	}
	unmatched, _ := localDefinitions(t, `schema: 1
site: unmatched
hosts: [example.com]
content:
  groups: [[article]]
  default_cleaner: false
`)
	snapshots["unmatched"] = unmatched
	source := `<title>Review article</title><article><h2>Verified reporting</h2>
<p>A substantive report includes a literal example: &lt;img src=x onerror=alert(1)&gt; and enough factual context to remain meaningful.</p>
<p>The concluding paragraph contains additional verified reporting and context, <em>emphasis &amp; &lt;literal&gt;</em> and <a href="https://example.com/reference">reference &amp; &lt;label&gt;</a>.</p>
<p>Literal entities: &amp;copy; &amp;lt;tag&amp;gt; and ordinary A &amp; B. Inline code: <code>&lt;img src=x onerror=alert(1)&gt; &amp;copy;</code>.</p>
<ul><li>First verified finding</li><li>Second verified finding</li></ul></article>`
	for name, snapshot := range snapshots {
		t.Run(name, func(t *testing.T) {
			options := []Option{WithContentType("markdown")}
			if name != "legacy-generic" {
				options = append(options, WithDefinitions(snapshot))
			}
			result, err := New(options...).ParseHTML(context.Background(), source, "https://93.184.216.34/review")
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{
				"A substantive report includes a literal example: &lt;img src=x onerror=alert(1)&gt; and enough factual context",
				"_emphasis &amp; &lt;literal&gt;_",
				"[reference &amp; &lt;label&gt;](https://example.com/reference)",
				"Literal entities: &amp;copy; &amp;lt;tag&amp;gt; and ordinary A &amp; B.",
				"`<img src=x onerror=alert(1)> &copy;`",
				"- First verified finding",
				"- Second verified finding",
			} {
				if !strings.Contains(result.Content, want) {
					t.Errorf("missing %q in:\n%s", want, result.Content)
				}
			}
			if name == "selected-false" && !strings.Contains(result.Content, "## Verified reporting") {
				t.Errorf("selected heading lost: %s", result.Content)
			}
		})
	}
}
