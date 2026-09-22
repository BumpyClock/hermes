package parser

import (
	"context"
	"fmt"

	"github.com/PuerkitoBio/goquery"

	"github.com/BumpyClock/hermes/internal/cleaners"
	"github.com/BumpyClock/hermes/internal/extractors"
	"github.com/BumpyClock/hermes/internal/utils/dom"
)

func processOrderedContent(ctx context.Context, elements *goquery.Selection, source *goquery.Document, extractor *extractors.ContentExtractor, title, targetURL, baseURL string) (string, error) {
	wrapper, err := extractor.OrderedTransforms.CopyAndExecute(ctx, outermostContentElements(elements), baseURL)
	if err != nil {
		return "", err
	}
	output := goquery.NewDocumentFromNode(wrapper.Get(0))
	for _, selector := range extractor.Clean {
		if err = ctx.Err(); err != nil {
			return "", err
		}
		wrapper.FindMatcher(selector).Remove()
	}
	output.Find("base").Remove()
	dom.MakeLinksAbsolute(output, baseURL)
	if !extractor.DisableDefaultCleaner {
		wrapper = cleaners.ExtractCleanNode(wrapper, source, cleaners.ContentCleanOptions{
			CleanConditionally: true, Title: title, URL: targetURL, PreserveMatchers: extractor.Preserve,
		})
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	content, err := wrapper.Html()
	if err != nil {
		return "", fmt.Errorf("serialize ordered content: %w", err)
	}
	return content, nil
}
