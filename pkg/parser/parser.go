package parser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Mercury is the main parser implementation
type Mercury struct {
	options ParserOptions
}

// New creates a new Mercury parser instance
func New(opts ...ParserOptions) *Mercury {
	parser := &Mercury{}
	if len(opts) > 0 {
		parser.options = opts[0]
	} else {
		parser.options = ParserOptions{
			FetchAllPages: true,
			Fallback:      true,
			ContentType:   "html",
		}
	}
	return parser
}

// Parse extracts content from a URL
func (m *Mercury) Parse(targetURL string, opts ParserOptions) (*Result, error) {
	// Parse and validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if !validateURL(parsedURL) {
		return &Result{
			Error:   true,
			Message: "The url parameter passed does not look like a valid URL. Please check your URL and try again.",
		}, nil
	}

	// TODO: Create resource (fetch or use provided HTML)
	// doc, err := resource.Create(targetURL, opts.Headers, "")
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create resource: %w", err)
	// }

	// TODO: Get appropriate extractor
	// extractor := extractors.GetExtractor(targetURL, parsedURL, doc)

	// For now, return a placeholder result
	result := &Result{
		URL:    targetURL,
		Domain: parsedURL.Host,
		Title:  "TODO: Implement extraction",
	}

	// TODO: Extract content
	// result, err := extractor.Extract(doc, targetURL, ExtractorOptions{
	//     URL:         targetURL,
	//     Fallback:    opts.Fallback,
	//     ContentType: opts.ContentType,
	// })

	// TODO: Handle multi-page articles if needed
	// if opts.FetchAllPages && result.NextPageURL != "" {
	//     result, err = m.collectAllPages(result, extractor, opts)
	//     if err != nil {
	//         return nil, fmt.Errorf("failed to collect pages: %w", err)
	//     }
	// }

	return result, nil
}

// ParseHTML extracts content from provided HTML
func (m *Mercury) ParseHTML(html string, targetURL string, opts ParserOptions) (*Result, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create document from HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// For now, return basic extraction
	result := &Result{
		URL:    targetURL,
		Domain: parsedURL.Host,
	}

	// Basic title extraction for testing
	title := doc.Find("title").First().Text()
	if title == "" {
		title = doc.Find("h1").First().Text()
	}
	result.Title = strings.TrimSpace(title)

	// Basic content extraction for testing
	content := doc.Find("article, .article, #article, .content, #content").First().Text()
	if content == "" {
		content = doc.Find("p").First().Text()
	}
	result.Content = strings.TrimSpace(content)

	// TODO: Continue with full extraction implementation
	// extractor := extractors.GetExtractor(targetURL, parsedURL, doc)
	// return extractor.Extract(doc, targetURL, ExtractorOptions{
	//     URL:         targetURL,
	//     HTML:        html,
	//     Fallback:    opts.Fallback,
	//     ContentType: opts.ContentType,
	// })

	return result, nil
}

func validateURL(u *url.URL) bool {
	return u.Scheme != "" && u.Host != ""
}

func (m *Mercury) collectAllPages(result *Result, extractor Extractor, opts ParserOptions) (*Result, error) {
	// TODO: Implementation for multi-page collection
	return result, nil
}