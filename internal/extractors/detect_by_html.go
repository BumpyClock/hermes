// ABOUTME: HTML-based extractor detection system that identifies extractors using CSS selectors
// ABOUTME: 100% JavaScript-compatible implementation of detect-by-html.js functionality

package extractors

import (
	"github.com/PuerkitoBio/goquery"

	"github.com/BumpyClock/hermes/internal/parser"
)

// Use the standard parser.Extractor interface.
type Extractor = parser.Extractor

// DetectByHTML identifies an appropriate extractor based on HTML meta tags
// JavaScript equivalent: export default function detectByHtml($).
func DetectByHTML(doc *goquery.Document) Extractor {
	// JavaScript logic:
	// const selector = Reflect.ownKeys(Detectors).find(s => $(s).length > 0);
	// return Detectors[selector];

	// Find the first selector that matches elements in the document
	for _, detector := range htmlDetectors {
		if doc.Find(detector.selector).Length() > 0 {
			return detector.extractor
		}
	}

	// Return nil if no detector matches
	return nil
}

// JavaScript equivalent: const Detectors = { ... }.
type htmlDetector struct {
	selector  string
	extractor Extractor
}

var htmlDetectors = [...]htmlDetector{
	// Match JavaScript selector exactly: 'meta[name="al:ios:app_name"][value="Medium"]'
	{selector: `meta[name="al:ios:app_name"][value="Medium"]`, extractor: &MediumExtractor{}},
	// Match JavaScript selector exactly: 'meta[name="generator"][value="blogger"]'
	{selector: `meta[name="generator"][value="blogger"]`, extractor: &BloggerExtractor{}},
}

// MediumExtractor represents the Medium.com custom extractor.
type MediumExtractor struct{}

func (m *MediumExtractor) GetDomain() string {
	return "medium.com"
}

func (m *MediumExtractor) Extract(doc *goquery.Document, url string, opts *parser.ExtractorOptions) (*parser.Result, error) {
	// Placeholder implementation - will be replaced with actual extraction logic
	return &parser.Result{
		URL:    url,
		Domain: "medium.com",
		Title:  "Medium Article (placeholder)",
	}, nil
}

// BloggerExtractor represents the Blogger/Blogspot custom extractor.
type BloggerExtractor struct{}

func (b *BloggerExtractor) GetDomain() string {
	return "blogspot.com"
}

func (b *BloggerExtractor) Extract(doc *goquery.Document, url string, opts *parser.ExtractorOptions) (*parser.Result, error) {
	// Placeholder implementation - will be replaced with actual extraction logic
	return &parser.Result{
		URL:    url,
		Domain: "blogspot.com",
		Title:  "Blogger Post (placeholder)",
	}, nil
}
