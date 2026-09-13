package parser

import "testing"

func TestMarkdownLiteralTextSerialization(t *testing.T) {
	for _, tc := range []struct {
		name, html, want string
	}{
		{"text", `<p>Example &lt;img src=x onerror=alert(1)&gt; &amp;copy; &amp;lt;tag&amp;gt; A &amp; B</p>`, `Example &lt;img src=x onerror=alert(1)&gt; &amp;copy; &amp;lt;tag&amp;gt; A &amp; B`},
		{"nested inline", `<h2>News &amp; context</h2><p><strong>Bold <em>&lt;example&gt;</em></strong> and <a href="https://example.com">A &amp; B</a></p>`, "## News &amp; context\n\n**Bold _&lt;example&gt;_** and [A &amp; B](https://example.com)"},
		{"inline code", "<p><code>&lt;tag&gt; &amp;copy; `x`</code></p>", "``<tag> &copy; `x` ``"},
		{"code block", "<pre><code>&lt;tag&gt; &amp;copy;\nA &amp; B</code></pre>", "```\n<tag> &copy;\nA & B\n```"},
		{"literal markdown", `<p>*literal* [label] &lt;tag&gt; \ &amp;copy;</p>`, `\*literal\* \[label\] &lt;tag&gt; \ &amp;copy;`},
		{"list", `<ul><li>A &amp; B</li><li>&lt;tag&gt;</li></ul>`, "- A &amp; B\n- &lt;tag&gt;"},
		{"nested code", `<p><code><span>&lt;tag&gt;</span> &amp;lt;tag&amp;gt;</code></p>`, "`<tag> &lt;tag&gt;`"},
		{"numeric entities", `<p>&amp;#60;tag&amp;#62; &amp;#x3c;tag&amp;#x3e;</p>`, `&amp;#60;tag&amp;#62; &amp;#x3c;tag&amp;#x3e;`},
		{"backslash", `<p>\&lt;tag&gt; \&amp;copy;</p>`, `\\&lt;tag&gt; \\&amp;copy;`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertToMarkdown(tc.html); got != tc.want {
				t.Errorf("markdown = %q, want %q", got, tc.want)
			}

		})
	}
}

func TestLiteralTextOtherFormatsUnchanged(t *testing.T) {
	const source = `<p>Meaningful &lt;img src=x onerror=alert(1)&gt; &amp;copy; A &amp; B</p>`
	if got := formatContent(source, "html"); got != source {
		t.Errorf("HTML = %q, want %q", got, source)
	}
	const wantText = `Meaningful <img src=x onerror=alert(1)> &copy; A & B`
	if got := formatContent(source, "text"); got != wantText {
		t.Errorf("text = %q, want %q", got, wantText)
	}
}
