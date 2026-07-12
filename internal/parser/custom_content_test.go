package parser

import (
	"errors"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"github.com/BumpyClock/hermes/internal/extractors/custom"
)

func TestParseHTMLUsesGothamistTransformsForImageCaptions(t *testing.T) {
	result, err := New().ParseHTML(`
		<html><body>
			<h1>Gothamist headline</h1>
			<div class="article-body">
				<p>Story text.</p>
				<div class="image-none"><img src="https://gothamist.com/photo.jpg" alt="Photo" width="640" height="360"><i>Photo caption</i></div>
			</div>
		</body></html>`, "https://gothamist.com/news/example", &ParserOptions{Fallback: true, ContentType: "html"})
	if err != nil {
		t.Fatal(err)
	}

	if result.ExtractorUsed != "custom:gothamist.com" {
		t.Fatalf("extractor used %q, want custom:gothamist.com", result.ExtractorUsed)
	}
	if !strings.Contains(result.Content, "<figure>") || !strings.Contains(result.Content, "<figcaption>Photo caption</figcaption>") {
		t.Fatalf("Gothamist image transform missing: %q", result.Content)
	}
	if !strings.Contains(result.Content, "<figure><img src=\"https://gothamist.com/photo.jpg\" alt=\"Photo\"/><figcaption>Photo caption</figcaption></figure>") {
		t.Fatalf("Gothamist image/caption nesting incorrect: %q", result.Content)
	}
}

func TestCustomExtractorPreservesRawContentSelectorFallbackOrder(t *testing.T) {
	result, err := New().ParseHTML(`
		<html><body>
			<article><p class="gallery">discarded()</p></article>
			<div itemprop="articleBody"><p>fallback content must not be selected</p></div>
		</body></html>`, "https://arstechnica.com/example", &ParserOptions{Fallback: false, ContentType: "html"})
	if err != nil {
		t.Fatal(err)
	}

	if result.ExtractorUsed != "custom:arstechnica.com" {
		t.Fatalf("extractor used %q, want custom:arstechnica.com", result.ExtractorUsed)
	}
	if strings.Contains(result.Content, "fallback content must not be selected") {
		t.Fatalf("later content selector replaced the first raw match: %q", result.Content)
	}
}

func TestContentElementsForSelectorGroupsPreserveSourceOrderAndDeduplicate(t *testing.T) {
	for name, selector := range map[string]interface{}{
		"string slice":    []string{".second", ".first", ".first"},
		"interface slice": []interface{}{".second", ".first", ".first"},
	} {
		t.Run(name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
				<html><body><article>
					<p class="first">first sibling</p>
					<p class="second">second sibling</p>
				</article></body></html>`))
			if err != nil {
				t.Fatal(err)
			}

			contentElements := contentElementsForSelector(doc, selector)
			if contentElements.Length() != 2 {
				t.Fatalf("content elements length = %d, want 2", contentElements.Length())
			}

			content, err := processCustomContent(
				contentElements,
				doc,
				&custom.ContentExtractor{DisableDefaultCleaner: true},
				"Title",
				"https://example.com/article",
			)
			if err != nil {
				t.Fatal(err)
			}

			if strings.Index(content, "first sibling") > strings.Index(content, "second sibling") {
				t.Fatalf("content order does not match source order: %q", content)
			}
		})
	}
}

func TestParseHTMLUsesStringSliceContentSelectorGroup(t *testing.T) {
	result, err := New().ParseHTML(`
		<html><body><section>
			<header><h1>Grouped selector headline</h1></header>
			<h2>Section heading</h2>
			<p>Grouped selector body.</p>
			<ol><li>Grouped selector list item.</li></ol>
		</section></body></html>`, "https://www.gruene.de/example", &ParserOptions{Fallback: false, ContentType: "html"})
	if err != nil {
		t.Fatal(err)
	}

	if result.ExtractorUsed != "custom:www.gruene.de" {
		t.Fatalf("extractor used %q, want custom:www.gruene.de", result.ExtractorUsed)
	}
	if !strings.Contains(result.Content, "Grouped selector body.") || !strings.Contains(result.Content, "Grouped selector list item.") {
		t.Fatalf("string-slice content selector group was not extracted: %q", result.Content)
	}
	if strings.Index(result.Content, "Section heading") > strings.Index(result.Content, "Grouped selector body.") {
		t.Fatalf("string-slice content selector group did not preserve source order: %q", result.Content)
	}
}

func TestProcessCustomContentAppliesTransformsBeforeCleanWithoutMutatingSource(t *testing.T) {
	removedTransformCalls := 0
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html><body><article>
			<div class="transform"><p>kept</p></div>
			<p class="remove">removed</p>
		</article></body></html>`))
	if err != nil {
		t.Fatal(err)
	}

	content, err := processCustomContent(
		doc.Find("article"),
		doc,
		&custom.ContentExtractor{
			DisableDefaultCleaner: true,
			Transforms: map[string]custom.TransformFunction{
				"article": &custom.FunctionTransform{Fn: func(selection *goquery.Selection) error {
					selection.AppendHtml(`<p data-transformed="true">transformed</p>`)
					return nil
				}},
				".remove": &custom.FunctionTransform{Fn: func(selection *goquery.Selection) error {
					removedTransformCalls++
					return nil
				}},
			},
			Clean: []string{".remove"},
		},
		"Title",
		"https://example.com/article",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(content, `<p data-transformed="true">transformed</p>`) {
		t.Fatalf("transform not applied: %q", content)
	}
	if strings.Contains(content, "removed") {
		t.Fatalf("clean selector not applied: %q", content)
	}
	if removedTransformCalls != 1 {
		t.Fatalf("transform did not run before clean: %d calls", removedTransformCalls)
	}
	if doc.Find("article [data-transformed]").Length() != 0 || doc.Find("article .remove").Length() != 1 {
		t.Fatalf("source DOM was mutated: %s", doc.Text())
	}
}

func TestProcessCustomContentCombinesMatchesAndAppliesDefaultCleaner(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html><body>
			<section class="part"><p>first</p><script>bad()</script><img src="spacer.gif" width="1" height="1"></section>
			<section class="part"><p><a href="/next">second</a></p></section>
		</body></html>`))
	if err != nil {
		t.Fatal(err)
	}

	content, err := processCustomContent(
		doc.Find(".part"),
		doc,
		&custom.ContentExtractor{},
		"Title",
		"https://example.com/article",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(content, "first") || !strings.Contains(content, "second") {
		t.Fatalf("matched elements not combined: %q", content)
	}
	if strings.Contains(content, "bad()") {
		t.Fatalf("default cleaner did not remove script: %q", content)
	}
	if strings.Contains(content, "spacer.gif") {
		t.Fatalf("default cleaner did not remove spacer image: %q", content)
	}
	if !strings.Contains(content, `href="https://example.com/next"`) {
		t.Fatalf("default cleaner did not resolve relative link: %q", content)
	}
	if doc.Find("script").Length() != 1 || doc.Find(`a[href="/next"]`).Length() != 1 {
		t.Fatal("default cleaner mutated source DOM")
	}
}

func TestProcessCustomContentSkipsNestedMatchesAndPreservesSiblings(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html><body>
			<article><p>main content</p><article><p>nested content</p></article></article>
			<article><p>sibling content</p></article>
		</body></html>`))
	if err != nil {
		t.Fatal(err)
	}

	content, err := processCustomContent(
		doc.Find("article"),
		doc,
		&custom.ContentExtractor{DisableDefaultCleaner: true},
		"Title",
		"https://example.com/article",
	)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(content, "main content") != 1 || strings.Count(content, "nested content") != 1 || strings.Count(content, "sibling content") != 1 {
		t.Fatalf("unexpected content selection: %q", content)
	}
	if strings.Index(content, "main content") > strings.Index(content, "sibling content") {
		t.Fatalf("sibling source order changed: %q", content)
	}
}

func TestProcessCustomContentCleanRemovesSelectedTopLevelElement(t *testing.T) {
	source := `<html><body><article><p>content</p></article></body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}

	content, err := processCustomContent(
		doc.Find("article"),
		doc,
		&custom.ContentExtractor{
			DisableDefaultCleaner: true,
			Clean:                 []string{"article"},
		},
		"Title",
		"https://example.com/article",
	)
	if err != nil {
		t.Fatal(err)
	}

	if content != "" {
		t.Fatalf("cleaned top-level element produced content: %q", content)
	}
	if doc.Find("article").Length() != 1 || doc.Find("article p").Text() != "content" {
		t.Fatalf("cleaning mutated source DOM: %q", doc.Text())
	}
	if !strings.Contains(source, "<article><p>content</p></article>") {
		t.Fatalf("source article/text was modified: %q", source)
	}
}

func TestProcessCustomContentSkipsDefaultCleanerWhenDisabled(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html><body><article><p>content</p><img src="spacer.gif" width="1" height="1"></article></body></html>`))
	if err != nil {
		t.Fatal(err)
	}

	content, err := processCustomContent(
		doc.Find("article"),
		doc,
		&custom.ContentExtractor{DisableDefaultCleaner: true},
		"Title",
		"https://example.com/article",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(content, "spacer.gif") {
		t.Fatalf("disabled default cleaner removed spacer image: %q", content)
	}
	if doc.Find(`img[src="spacer.gif"]`).Length() != 1 {
		t.Fatal("disabled default cleaner mutated source DOM")
	}
}

func TestProcessCustomContentTransformsDescendantsBeforeParents(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html><body><article>
			<div class="image-none"><i>Ada Example</i></div>
		</article></body></html>`))
	if err != nil {
		t.Fatal(err)
	}

	content, err := processCustomContent(
		doc.Find("article"),
		doc,
		&custom.ContentExtractor{Transforms: map[string]custom.TransformFunction{
			"div.image-none": &custom.StringTransform{TargetTag: "figure"},
			".image-none i":  &custom.StringTransform{TargetTag: "figcaption"},
		}},
		"Title",
		"https://example.com/article",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(content, "<figure><figcaption>Ada Example</figcaption></figure>") {
		t.Fatalf("nested transforms ran in wrong order: %q", content)
	}
}

func TestProcessCustomContentDeduplicatesOverlappingTransformMatches(t *testing.T) {
	secondTransformCalls := 0
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html><body><article><div class="item"><p>content</p></div></article></body></html>`))
	if err != nil {
		t.Fatal(err)
	}

	content, err := processCustomContent(
		doc.Find("article"),
		doc,
		&custom.ContentExtractor{Transforms: map[string]custom.TransformFunction{
			".item": &custom.StringTransform{TargetTag: "figure"},
			"div.item": &custom.FunctionTransform{Fn: func(*goquery.Selection) error {
				secondTransformCalls++
				return nil
			}},
		}},
		"Title",
		"https://example.com/article",
	)
	if err != nil {
		t.Fatal(err)
	}

	if secondTransformCalls != 0 {
		t.Fatalf("overlapping transform ran %d times", secondTransformCalls)
	}
	if !strings.Contains(content, "<figure><p>content</p></figure>") {
		t.Fatalf("overlapping transform lost content: %q", content)
	}
}

func TestProcessCustomContentPropagatesTransformErrors(t *testing.T) {
	sentinel := errors.New("transform failed")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html><body><article><p>content</p></article></body></html>`))
	if err != nil {
		t.Fatal(err)
	}

	_, err = processCustomContent(
		doc.Find("article"),
		doc,
		&custom.ContentExtractor{Transforms: map[string]custom.TransformFunction{
			"p": &custom.FunctionTransform{Fn: func(*goquery.Selection) error {
				return sentinel
			}},
		}},
		"Title",
		"https://example.com/article",
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
