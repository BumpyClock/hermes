package hermes

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"golang.org/x/net/html"
)

func TestContentTypeMarkdownBehavior(t *testing.T) {
	// This test verifies that ContentType: "markdown" only affects the Content field,
	// while all other fields remain as properly typed values.
	client := New(WithContentType("markdown"))

	// Test HTML with various content types
	html := `
	<html>
		<head>
			<title>Test Article Title</title>
			<meta name="author" content="John Doe">
			<meta name="description" content="This is a test article">
		</head>
		<body>
			<article>
				<header>
					<h1>Main Heading</h1>
					<p class="author">By John Doe</p>
				</header>
				<div class="content">
					<p>This is the first paragraph with <strong>bold text</strong>.</p>
					<p>This is the second paragraph with <em>italic text</em>.</p>
					<ul>
						<li>List item 1</li>
						<li>List item 2</li>
					</ul>
				</div>
			</article>
		</body>
	</html>
	`

	result, err := client.ParseHTML(context.Background(), html, "https://example.com/test")
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	// Test that metadata fields are properly typed (not markdown)
	if result.Title == "" {
		t.Error("Title should not be empty")
	}
	if strings.Contains(result.Title, "#") || strings.Contains(result.Title, "**") {
		t.Errorf("Title should not contain markdown formatting: %s", result.Title)
	}

	// Author might be empty with generic extraction - that's OK
	if result.Author != "" && (strings.Contains(result.Author, "#") || strings.Contains(result.Author, "**")) {
		t.Errorf("Author should not contain markdown formatting: %s", result.Author)
	}

	if result.Description == "" {
		t.Error("Description should not be empty")
	}
	if strings.Contains(result.Description, "#") || strings.Contains(result.Description, "**") {
		t.Errorf("Description should not contain markdown formatting: %s", result.Description)
	}

	// Test that Content field IS in markdown format
	if result.Content == "" {
		t.Error("Content should not be empty")
	}

	// Content should contain markdown formatting
	hasMarkdown := strings.Contains(result.Content, "**") || strings.Contains(result.Content, "*") || strings.Contains(result.Content, "#")
	if !hasMarkdown {
		t.Errorf("Content should contain markdown formatting but got: %s", result.Content)
	}

	// Test that other fields are properly typed
	if result.WordCount <= 0 {
		t.Error("WordCount should be positive")
	}

	// Print results for manual verification
	t.Logf("Title: %q (type: %T)", result.Title, result.Title)
	t.Logf("Author: %q (type: %T)", result.Author, result.Author)
	t.Logf("Description: %q (type: %T)", result.Description, result.Description)
	t.Logf("Content: %q (type: %T)", result.Content, result.Content)
	t.Logf("WordCount: %d (type: %T)", result.WordCount, result.WordCount)

	// Test FormatMarkdown separately
	formatted := result.FormatMarkdown()
	if formatted == "" {
		t.Error("FormatMarkdown should not return empty string")
	}

	// FormatMarkdown should combine everything into one markdown document
	if !strings.Contains(formatted, "# "+result.Title) {
		t.Error("FormatMarkdown should contain title as H1")
	}

	t.Logf("FormatMarkdown output length: %d", len(formatted))
}

func TestMarkdownPreservesSpaceAfterCitationLink(t *testing.T) {
	client := New(WithContentType("markdown"))
	source := `<article><p><sup><a href="#citation"><span>Citation</span></a></sup> <a href="/next">Following link</a> supplies enough article detail for generic extraction.</p></article>`

	result, err := client.ParseHTML(context.Background(), source, "https://93.184.216.34/article")
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	const want = "[Citation](https://93.184.216.34/article#citation) [Following link](https://93.184.216.34/next) supplies enough article detail for generic extraction."
	if result.Content != want {
		t.Errorf("Markdown content = %q, want %q", result.Content, want)
	}
}

func TestMarkdownCitationLinkBoundaryMatrix(t *testing.T) {
	for _, tc := range []struct {
		name, boundary, want string
	}{
		{name: "span-em-nested-citation", boundary: `<span><em><sup><a href="#citation"><span>Citation</span></a></sup></em></span> <a href="/next">Following link</a>`, want: `_[Citation](https://93.184.216.34/article#citation)_ [Following link](https://93.184.216.34/next)`},
		{name: "tab", boundary: `<sup><a href="#citation"><span>Citation</span></a></sup>` + "\t" + `<a href="/next">Following link</a>`, want: `[Citation](https://93.184.216.34/article#citation) [Following link](https://93.184.216.34/next)`},
		{name: "newline", boundary: `<sup><a href="#citation"><span>Citation</span></a></sup>` + "\n" + `<a href="/next">Following link</a>`, want: `[Citation](https://93.184.216.34/article#citation) [Following link](https://93.184.216.34/next)`},
		{name: "no-whitespace", boundary: `<sup><a href="#citation"><span>Citation</span></a></sup><a href="/next">Following link</a>`, want: `[Citation](https://93.184.216.34/article#citation)[Following link](https://93.184.216.34/next)`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := `<article><p>` + tc.boundary + ` supplies enough article detail for generic extraction.</p></article>`
			result, err := New(WithContentType("markdown")).ParseHTML(context.Background(), source, "https://93.184.216.34/article")
			if err != nil {
				t.Fatalf("ParseHTML failed: %v", err)
			}
			want := tc.want + " supplies enough article detail for generic extraction."
			if result.Content != want {
				t.Errorf("Markdown content = %q, want %q", result.Content, want)
			}
		})
	}
}

func TestMarkdownLinkFormattingPreservesInlineContent(t *testing.T) {
	source := `<article><p>See <a href="/detail"><em>Emphasized</em> and <code>literal</code></a> after this sufficiently detailed article content.</p></article>`
	result, err := New(WithContentType("markdown")).ParseHTML(context.Background(), source, "https://93.184.216.34/article")
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	const want = "See [_Emphasized_ and `literal`](https://93.184.216.34/detail) after this sufficiently detailed article content."
	if result.Content != want {
		t.Errorf("Markdown content = %q, want %q", result.Content, want)
	}
}

func TestMarkdownPreservesInlineWrapperWhitespace(t *testing.T) {
	source := `<article><p><sup><a href="#citation"><span>Citation</span></a>: <span>Detail</span>, <span>More</span> <span>Text</span></sup> concludes this sufficiently detailed article content.</p></article>`
	result, err := New(WithContentType("markdown")).ParseHTML(context.Background(), source, "https://93.184.216.34/article")
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	const want = "[Citation](https://93.184.216.34/article#citation): Detail, More Text concludes this sufficiently detailed article content."
	if result.Content != want {
		t.Errorf("Markdown content = %q, want %q", result.Content, want)
	}
}

func TestMarkdownPreservesLinkLabelBoundaryWhitespace(t *testing.T) {
	for _, tc := range []struct {
		name, label, want string
	}{
		{name: "leading-space", label: " label", want: "prefix [label](https://93.184.216.34/detail)suffix"},
		{name: "trailing-space", label: "label ", want: "prefix[label](https://93.184.216.34/detail) suffix"},
		{name: "both-spaces", label: " label ", want: "prefix [label](https://93.184.216.34/detail) suffix"},
		{name: "leading-tab", label: "\tlabel", want: "prefix [label](https://93.184.216.34/detail)suffix"},
		{name: "trailing-newline", label: "label\n", want: "prefix[label](https://93.184.216.34/detail) suffix"},
		{name: "empty-label", label: " ", want: "prefix suffix"},
		{name: "adjacent", label: "label", want: "prefix[label](https://93.184.216.34/detail)suffix"},
		{name: "leading-non-breaking-space", label: "\u00a0label", want: "prefix[\u00a0label](https://93.184.216.34/detail)suffix"},
		{name: "trailing-non-breaking-space", label: "label\u00a0", want: "prefix[label\u00a0](https://93.184.216.34/detail)suffix"},
		{name: "mixed-ascii-and-non-breaking-space", label: " \u00a0label", want: "prefix [\u00a0label](https://93.184.216.34/detail)suffix"},
		{name: "leading-narrow-non-breaking-space", label: "\u202flabel", want: "prefix[\u202flabel](https://93.184.216.34/detail)suffix"},
		{name: "leading-ideographic-space", label: "\u3000label", want: "prefix[\u3000label](https://93.184.216.34/detail)suffix"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := `<article><p>prefix<a href="/detail">` + tc.label + `</a>suffix with enough detailed article content for generic extraction.</p></article>`
			result, err := New(WithContentType("markdown")).ParseHTML(context.Background(), source, "https://93.184.216.34/article")
			if err != nil {
				t.Fatalf("ParseHTML failed: %v", err)
			}
			want := tc.want + " with enough detailed article content for generic extraction."
			if result.Content != want {
				t.Errorf("Markdown content = %q, want %q", result.Content, want)
			}
		})
	}
}

func TestMarkdownPreservesJapaneseStrongAdjacency(t *testing.T) {
	source := `<article><p>日本<strong>語</strong>！これは詳細な記事本文で、強調された語の直後に句読点が続きます。</p></article>`
	result, err := New(WithContentType("markdown")).ParseHTML(context.Background(), source, "https://93.184.216.34/article")
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}

	const wantMarkdown = "日本**語**！これは詳細な記事本文で、強調された語の直後に句読点が続きます。"
	if result.Content != wantMarkdown {
		t.Errorf("Markdown content = %q, want %q", result.Content, wantMarkdown)
	}
	const wantVisible = "日本語！これは詳細な記事本文で、強調された語の直後に句読点が続きます。"
	if got := renderedMarkdownParagraphText(t, result.Content); got != wantVisible {
		t.Errorf("Rendered visible text = %q, want %q", got, wantVisible)
	}
}

func TestMarkdownPreservesBracketedEmphasisAdjacency(t *testing.T) {
	for _, tc := range []struct {
		source, wantMarkdown, wantVisible string
	}{
		{`<p>prefix [<i>Title</i>] suffix with enough detailed article content for generic extraction.</p>`, `prefix \[*Title*\] suffix with enough detailed article content for generic extraction.`, "prefix [Title] suffix with enough detailed article content for generic extraction."},
		{`<p>[<em>Title</em>], next with enough detailed article content for generic extraction.</p>`, `\[*Title*\], next with enough detailed article content for generic extraction.`, "[Title], next with enough detailed article content for generic extraction."},
	} {
		result, err := New(WithContentType("markdown")).ParseHTML(context.Background(), `<article>`+tc.source+`</article>`, "https://93.184.216.34/article")
		if err != nil {
			t.Fatalf("ParseHTML failed: %v", err)
		}
		if result.Content != tc.wantMarkdown {
			t.Errorf("Markdown content = %q, want %q", result.Content, tc.wantMarkdown)
		}
		if got := renderedMarkdownParagraphText(t, result.Content); got != tc.wantVisible {
			t.Errorf("Rendered visible text = %q, want %q", got, tc.wantVisible)
		}
	}

}

func TestConfiguredMarkdownTableBoundaries(t *testing.T) {
	client := markdownSnapshotClient(t)
	source := `<article><table><caption>Table caption</caption><tr><th>Column A</th><th>Column B</th></tr><tr><td>Cell A</td><td>Cell B</td></tr></table></article>`
	result, err := client.ParseHTML(context.Background(), source, "https://93.184.216.34/article")
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}
	const wantMarkdown = "Table caption\n\nColumn A\nColumn B\n\nCell A\nCell B"
	if result.Content != wantMarkdown {
		t.Errorf("Markdown content = %q, want %q", result.Content, wantMarkdown)
	}
	if got := renderedMarkdownVisibleText(t, result.Content); got != "Table caption Column A Column B Cell A Cell B" {
		t.Errorf("Rendered visible text = %q", got)
	}
}

func TestMarkdownPreservesLinkedEmphasisWrapperWhitespace(t *testing.T) {
	source := `<article><p><em>Read </em><a href="/next"><em>here.</em></a> with enough detailed article content for generic extraction.</p></article>`
	result, err := New(WithContentType("markdown")).ParseHTML(context.Background(), source, "https://93.184.216.34/article")
	if err != nil {
		t.Fatalf("ParseHTML failed: %v", err)
	}
	const wantMarkdown = `_Read_ [_here._](https://93.184.216.34/next) with enough detailed article content for generic extraction.`
	if result.Content != wantMarkdown {
		t.Errorf("Markdown content = %q, want %q", result.Content, wantMarkdown)
	}
	const wantVisible = "Read here. with enough detailed article content for generic extraction."
	if got := renderedMarkdownParagraphText(t, result.Content); got != wantVisible {
		t.Errorf("Rendered visible text = %q, want %q", got, wantVisible)
	}
}

func markdownSnapshotClient(t *testing.T) *Client {
	t.Helper()
	directory := t.TempDir()
	definition := []byte("schema: 1\n" +
		"site: example\n" +
		"hosts: [93.184.216.34]\n" +
		"content:\n" +
		"  groups: [[article]]\n" +
		"  default_cleaner: false\n")
	if err := os.WriteFile(filepath.Join(directory, "example.yaml"), definition, 0o600); err != nil {
		t.Fatalf("write definition: %v", err)
	}
	definitions, err := LoadDefinitions(directory)
	if err != nil {
		t.Fatalf("load definitions: %v", err)
	}
	return New(WithDefinitions(definitions), WithContentType("markdown"))
}

func renderedMarkdownParagraphText(t *testing.T, markdown string) string {
	t.Helper()
	var rendered bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &rendered); err != nil {
		t.Fatalf("render Markdown: %v", err)
	}
	document, err := html.Parse(&rendered)
	if err != nil {
		t.Fatalf("parse rendered Markdown: %v", err)
	}
	paragraph := firstElement(document, "p")
	if paragraph == nil {
		t.Fatal("rendered Markdown has no paragraph")
	}
	var text strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(paragraph)
	return text.String()
}

func renderedMarkdownVisibleText(t *testing.T, markdown string) string {
	t.Helper()
	var rendered bytes.Buffer
	if err := goldmark.Convert([]byte(markdown), &rendered); err != nil {
		t.Fatalf("render Markdown: %v", err)
	}
	document, err := html.Parse(&rendered)
	if err != nil {
		t.Fatalf("parse rendered Markdown: %v", err)
	}
	var text strings.Builder
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
			text.WriteByte(' ')
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)
	return strings.Join(strings.Fields(text.String()), " ")
}

func firstElement(node *html.Node, name string) *html.Node {
	if node.Type == html.ElementNode && node.Data == name {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := firstElement(child, name); found != nil {
			return found
		}
	}
	return nil
}
