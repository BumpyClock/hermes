package parser

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"github.com/BumpyClock/hermes/internal/cleaners"
	"github.com/BumpyClock/hermes/internal/extractors"
	"github.com/BumpyClock/hermes/internal/utils/dom"
)

// processDefinitionContent cleans and serializes a rule's matched content.
// Rules without transforms clean each outermost match on its own and emit its
// inner HTML; transform rules clean the whole transformed copy once and emit
// the matches' outer HTML. The two shapes produce different output, so an
// empty transform program cannot stand in for the per-element path.
func processDefinitionContent(ctx context.Context, elements *goquery.Selection, source *goquery.Document, extractor *extractors.ContentExtractor, title, targetURL, baseURL string) (string, error) {
	elements = outermostContentElements(elements)
	if extractor.OrderedTransforms == nil {
		return processElementContent(ctx, elements, source, extractor, title, targetURL, baseURL)
	}
	return processOrderedContent(ctx, elements, source, extractor, title, targetURL, baseURL)
}

func processElementContent(ctx context.Context, elements *goquery.Selection, source *goquery.Document, extractor *extractors.ContentExtractor, title, targetURL, baseURL string) (string, error) {
	var combined strings.Builder
	for _, node := range elements.Nodes {
		// The parsed document gives each copy html and body ancestors, which
		// preserve selectors can match.
		contentDoc, err := goquery.NewDocumentFromReader(strings.NewReader("<div></div>"))
		if err != nil {
			return "", fmt.Errorf("create custom content wrapper: %w", err)
		}
		wrapper := contentDoc.Find("div").First()
		wrapper.AppendSelection(goquery.NewDocumentFromNode(node).Clone())
		if err = prepareDefinitionContent(ctx, wrapper, extractor, baseURL); err != nil {
			return "", err
		}
		content := cleanDefinitionContent(wrapper.Children().First(), source, extractor, title, targetURL)
		serialized, err := content.Html()
		if err != nil {
			return "", fmt.Errorf("serialize custom content: %w", err)
		}
		if strings.TrimSpace(serialized) != "" {
			combined.WriteString(serialized)
			combined.WriteByte('\n')
		}
	}
	return strings.TrimSpace(combined.String()), nil
}

func processOrderedContent(ctx context.Context, elements *goquery.Selection, source *goquery.Document, extractor *extractors.ContentExtractor, title, targetURL, baseURL string) (string, error) {
	wrapper, err := extractor.OrderedTransforms.CopyAndExecute(ctx, elements, baseURL)
	if err != nil {
		return "", err
	}
	if err = prepareDefinitionContent(ctx, wrapper, extractor, baseURL); err != nil {
		return "", err
	}
	wrapper = cleanDefinitionContent(wrapper, source, extractor, title, targetURL)
	if err = ctx.Err(); err != nil {
		return "", err
	}
	content, err := wrapper.Html()
	if err != nil {
		return "", fmt.Errorf("serialize ordered content: %w", err)
	}
	return content, nil
}

// prepareDefinitionContent applies the rule's removals to a detached copy.
// Rules see source attributes; only the copy receives absolute URLs.
func prepareDefinitionContent(ctx context.Context, wrapper *goquery.Selection, extractor *extractors.ContentExtractor, baseURL string) error {
	for _, selector := range extractor.Clean {
		if err := ctx.Err(); err != nil {
			return err
		}
		wrapper.FindMatcher(selector).Remove()
	}
	output := goquery.NewDocumentFromNode(wrapper.Get(0))
	output.Find("base").Remove()
	dom.MakeLinksAbsolute(output, baseURL)
	return nil
}

func cleanDefinitionContent(content *goquery.Selection, source *goquery.Document, extractor *extractors.ContentExtractor, title, targetURL string) *goquery.Selection {
	if extractor.DisableDefaultCleaner {
		return content
	}
	return cleaners.ExtractCleanNode(content, source, cleaners.ContentCleanOptions{
		CleanConditionally: true, Title: title, URL: targetURL, PreserveMatchers: extractor.Preserve,
	})
}

func outermostContentElements(contentElements *goquery.Selection) *goquery.Selection {
	selected := make(map[*html.Node]struct{}, contentElements.Length())
	for _, node := range contentElements.Nodes {
		selected[node] = struct{}{}
	}

	return contentElements.FilterFunction(func(_ int, element *goquery.Selection) bool {
		for ancestor := element.Get(0).Parent; ancestor != nil; ancestor = ancestor.Parent {
			if _, ok := selected[ancestor]; ok {
				return false
			}
		}
		return true
	})
}
