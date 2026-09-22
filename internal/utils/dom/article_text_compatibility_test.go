package dom

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func TestArticleTextExactCompatibility(t *testing.T) {
	for _, source := range articleTextInputs() {
		if got, want := StripTagsWithBlockBoundaries(source), originalArticleText(source); got != want {
			t.Fatalf("source=%q:\ngot  %q\nwant %q", source, got, want)
		}
	}
}

func articleTextInputs() []string {
	return []string{
		"", "literal &amp; text", " \u00a0literal \t text",
		`<!doctype html><head><title>Hidden</title></head><body><p>One</p><!--ignored--><p>Two</p></body>`,
		`<p>pre<span>fix</span>&nbsp;<strong>word</strong><br>next</p>`,
		`<article> before <template><p>hidden</p></template> after <noscript>hidden</noscript> tail </article>`,
		`<table>foster text<tr><td>First<td>Second<tr><td>Third</table>`,
		`<p>before<div>middle</p>after</div><p>end`,
		`<svg><title>SVG title</title><text>SVG text</text></svg><math><mi>x</mi></math>`,
		`<pre>  first&#10;  second</pre><code>&lt;literal&gt;</code><textarea>textarea text</textarea>`,
		"<div>\xfftext\xfe<script>hidden</script>\v</div>",
	}
}

func originalArticleText(source string) string {
	if strings.IndexByte(source, '<') == -1 {
		return source
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(source))
	if err != nil {
		return source
	}
	doc.Find("script, style, noscript, head, meta, link").Remove()
	doc.Find("template").Remove()
	removeComments(doc.Nodes[0])
	var output strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		block := false
		if node.Type == html.ElementNode {
			switch node.Data {
			case "address", "article", "aside", "blockquote", "br", "caption", "dd", "div",
				"dl", "dt", "fieldset", "figcaption", "figure", "footer", "form",
				"h1", "h2", "h3", "h4", "h5", "h6", "header", "hr", "li",
				"main", "nav", "ol", "p", "pre", "section", "table", "tbody", "td",
				"tfoot", "th", "thead", "tr", "ul":
				block = true
			}
		}
		if block {
			output.WriteByte(' ')
		}
		if node.Type == html.TextNode {
			output.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
		if block {
			output.WriteByte(' ')
		}
	}
	for _, node := range doc.Nodes {
		walk(node)
	}
	return output.String()
}

func FuzzArticleTextCompatibility(f *testing.F) {
	for _, source := range articleTextInputs() {
		f.Add(source)
	}
	f.Fuzz(func(t *testing.T, source string) {
		if len(source) > 8192 {
			t.Skip()
		}
		if got, want := StripTagsWithBlockBoundaries(source), originalArticleText(source); got != want {
			t.Fatalf("source=%q:\ngot  %q\nwant %q", source, got, want)
		}
	})
}

func BenchmarkArticleTextTraversal(b *testing.B) {
	source := "<article>" + strings.Repeat(`<section><h2>Heading</h2><p>Article <strong>words</strong>.</p><!--ignored--><template>hidden</template></section>`, 100) + "</article>"
	b.Run("original", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = originalArticleText(source)
		}
	})
	b.Run("single-walk", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = StripTagsWithBlockBoundaries(source)
		}
	})
}
