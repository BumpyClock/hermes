package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"golang.org/x/net/html"
)

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

func TestMarkdownLinkTitleSerialization(t *testing.T) {
	for _, tc := range []struct {
		name, html, want string
	}{
		{"terminal backslash", `<p><a href="https://example.test/detail" title="ends \">label</a></p>`, `[label](https://example.test/detail "ends \\")`},
		{"backslash before quote", `<p><a href="https://example.test/detail" title="slash \&quot;quote">label</a></p>`, `[label](https://example.test/detail "slash \\\"quote")`},
		{"multiline", "<p><a href=\"https://example.test/detail\" title=\"line one\r\nline two\">label</a></p>", `[label](https://example.test/detail "line one line two")`},
		{"parenthesized destination", `<p><a href="https://example.test/detail_(safe)" title="title">label</a></p>`, `[label](https://example.test/detail_(safe) "title")`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertToMarkdown(tc.html); got != tc.want {
				t.Errorf("markdown = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestJapaneseStrongAdjacencySerialization(t *testing.T) {
	for _, tc := range []struct {
		name, source, wantMarkdown, wantVisible string
	}{
		{"Japanese word boundary", `<p>日本<strong>語</strong>！</p>`, `日本**語**！`, "日本語！"},
		{"Japanese punctuation boundary", `<p>。<strong>語</strong>！</p>`, `。**語**！`, "。語！"},
		{"Japanese leading boundary", `<p><strong>語</strong>の</p>`, `**語**の`, "語の"},
		{"ASCII word boundary", `<p>prefix<strong>bold</strong>suffix</p>`, `prefix**bold**suffix`, "prefixboldsuffix"},
		{"Japanese quote boundary falls through", `<p>前<strong>「語」</strong>後</p>`, `前 **「語」** 後`, "前 「語」 後"},
		{"Japanese quote needs opening padding", `<p>前<strong>「語」で</strong>、</p>`, `前 **「語」で**、`, "前 「語」で、"},
		{"parenthesized content needs both padding", `<p>号<strong>(語)</strong>を</p>`, `号 **(語)** を`, "号 (語) を"},
		{"Japanese comma needs trailing padding", `<p>の<strong>運、</strong>発</p>`, `の**運、** 発`, "の運、 発"},
		{"nested emphasis keeps existing padding", `<p>日本<strong><em>語</em></strong>！</p>`, `日本 **_語_**！`, "日本 語！"},
		{"edge whitespace preserves source spacing", `<p>日本<strong> 語 </strong>！</p>`, `日本 **語** ！`, "日本 語 ！"},
		{"multiline keeps existing padding", "<p>日本<strong>語\n語</strong>！</p>", "日本**語 語**！", "日本語 語！"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertToMarkdown(tc.source); got != tc.wantMarkdown {
				t.Errorf("markdown = %q, want %q", got, tc.wantMarkdown)
			}
			if rendered := renderMarkdown(tc.source); rendered.visible != tc.wantVisible || !rendered.strong {
				t.Errorf("rendered = %+v, want visible %q with strong semantics", rendered, tc.wantVisible)
			}
		})
	}
}

func TestJapaneseStrongAdjacencyRendersSemantically(t *testing.T) {
	rendered := renderMarkdown(`<p>日本<strong>語</strong>！</p>`)
	if !rendered.strong {
		t.Fatal("rendered Markdown does not retain strong semantics")
	}
	if rendered.visible != "日本語！" {
		t.Errorf("rendered visible text = %q, want %q", rendered.visible, "日本語！")
	}
}

type markdownRender struct {
	visible string
	strong  bool
	em      bool
}

func renderMarkdown(source string) markdownRender {
	markdown := convertToMarkdown(source)
	var rendered bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &rendered); err != nil {
		panic(err)
	}
	document, err := html.Parse(&rendered)
	if err != nil {
		panic(err)
	}
	return markdownRender{
		visible: strings.TrimSpace(documentText(document)),
		strong:  documentContainsTag(document, "strong"),
		em:      documentContainsTag(document, "em"),
	}
}

func TestAdjacentEmphasisSerialization(t *testing.T) {
	for _, tc := range []struct {
		name, source, wantMarkdown, wantVisible string
	}{
		{"bracketed italic", `<p>prefix [<i>Title</i>] suffix</p>`, `prefix \[*Title*\] suffix`, "prefix [Title] suffix"},
		{"bracketed emphasis", `<p>[<em>Title</em>], next</p>`, `\[*Title*\], next`, "[Title], next"},
		{"Japanese quote boundary falls through", `<p>前<i>「語」</i>後</p>`, `前 _「語」_ 後`, "前 「語」 後"},
		{"nested strong keeps existing behavior", `<p>[<em><strong>Title</strong></em>], next</p>`, `\[ _**Title**_\], next`, "[ Title], next"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertToMarkdown(tc.source); got != tc.wantMarkdown {
				t.Errorf("markdown = %q, want %q", got, tc.wantMarkdown)
			}
			rendered := renderMarkdown(tc.source)
			if rendered.visible != tc.wantVisible || !rendered.em {
				t.Errorf("rendered = %+v, want visible %q with emphasis semantics", rendered, tc.wantVisible)
			}
		})
	}
}

func TestCommonMarkSymbolFlankingFallsBackToPaddedMarkup(t *testing.T) {
	for _, tag := range []struct {
		name, open, close string
	}{
		{"strong", "<strong>", "</strong>"},
		{"emphasis", "<em>", "</em>"},
	} {
		for _, symbol := range "$+<=>^`|~" {
			t.Run(tag.name+"-"+string(symbol), func(t *testing.T) {
				source := "<p>a" + tag.open + string(symbol) + "5" + tag.close + "b</p>"
				markdown := convertToMarkdown(source)
				escapedSymbol := string(symbol)
				switch symbol {
				case '<':
					escapedSymbol = "&lt;"
				case '>':
					escapedSymbol = "&gt;"
				case '`', '|':
					escapedSymbol = `\` + escapedSymbol
				}
				var want string
				if tag.name == "strong" {
					want = "a **" + escapedSymbol + "5** b"
				} else {
					want = "a _" + escapedSymbol + "5_ b"
				}
				if markdown != want {
					t.Errorf("markdown = %q, want %q", markdown, want)
				}
				rendered := renderMarkdown(source)
				if rendered.visible != "a "+string(symbol)+"5 b" || !rendered.strong && tag.name == "strong" || !rendered.em && tag.name == "emphasis" {
					t.Errorf("rendered = %+v, want semantic %s", rendered, tag.name)
				}
			})
		}
	}

}

func TestMarkdownTableAndCaptionBoundaries(t *testing.T) {
	source := `<figure><img src="https://example.test/photo.jpg" alt="Photo"><figcaption>Figure caption</figcaption></figure><table><caption>Table caption</caption><tr><th>Column <em>A</em></th><th>Column B</th></tr><tr><td>Cell A</td><td>Cell B</td></tr></table>`
	const wantMarkdown = "![Photo](https://example.test/photo.jpg)\n\nFigure caption\n\nTable caption\n\nColumn _A_\nColumn B\n\nCell A\nCell B"
	if got := convertToMarkdown(source); got != wantMarkdown {
		t.Errorf("markdown = %q, want %q", got, wantMarkdown)
	}

	rendered := renderMarkdown(source)
	if got := strings.Join(strings.Fields(rendered.visible), " "); got != "Figure caption Table caption Column A Column B Cell A Cell B" {
		t.Errorf("rendered visible text = %q", rendered.visible)
	}
	if !rendered.em {
		t.Error("rendered table cell does not retain emphasis semantics")
	}
}

func TestInlineWrapperWhitespaceSerialization(t *testing.T) {
	for _, tc := range []struct {
		name, source, wantMarkdown, wantVisible string
	}{
		{"trailing ASCII before linked emphasis", `<p><em>Read </em><a href="/next"><em>here.</em></a></p>`, `_Read_ [_here._](/next)`, "Read here."},
		{"leading ASCII", `<p>prefix<em> Read</em> suffix</p>`, `prefix _Read_ suffix`, "prefix Read suffix"},
		{"non-breaking spaces", "<p>prefix<em>\u00a0Read\u00a0</em>suffix</p>", "prefix&#160;_Read_&#160;suffix", "prefix\u00a0Read\u00a0suffix"},
		{"narrow and ideographic spaces", "<p>prefix<em>\u202fRead\u3000</em>suffix</p>", "prefix&#8239;_Read_&#12288;suffix", "prefix\u202fRead\u3000suffix"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertToMarkdown(tc.source); got != tc.wantMarkdown {
				t.Errorf("markdown = %q, want %q", got, tc.wantMarkdown)
			}
			rendered := renderMarkdown(tc.source)
			if rendered.visible != tc.wantVisible || !rendered.em {
				t.Errorf("rendered = %+v, want visible %q with emphasis semantics", rendered, tc.wantVisible)
			}
		})
	}
}

func documentContainsTag(node *html.Node, tag string) bool {
	if node.Type == html.ElementNode && node.Data == tag {
		return true
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if documentContainsTag(child, tag) {
			return true
		}
	}
	return false
}

func documentText(node *html.Node) string {
	var text strings.Builder
	var visit func(*html.Node)
	visit = func(current *html.Node) {
		if current.Type == html.TextNode {
			text.WriteString(current.Data)
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(node)
	return text.String()
}
