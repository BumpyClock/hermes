// ABOUTME: Custom extractor interface and types for site-specific content extraction
// ABOUTME: Foundation structure for 150+ domain-specific extractors with transforms, selectors, and cleaning

package custom

import (
	"github.com/PuerkitoBio/goquery"
)

// CustomExtractor represents a site-specific content extractor
// JavaScript equivalent: Each extractor export in custom/[domain]/index.js.
type CustomExtractor struct {
	Domain           string                     `json:"domain"`
	SupportedDomains []string                   `json:"supportedDomains,omitempty"`
	Title            *FieldExtractor            `json:"title,omitempty"`
	Author           *FieldExtractor            `json:"author,omitempty"`
	Content          *ContentExtractor          `json:"content,omitempty"`
	DatePublished    *FieldExtractor            `json:"date_published,omitempty"`
	LeadImageURL     *FieldExtractor            `json:"lead_image_url,omitempty"`
	Dek              *FieldExtractor            `json:"dek,omitempty"`
	NextPageURL      *FieldExtractor            `json:"next_page_url,omitempty"`
	Excerpt          *FieldExtractor            `json:"excerpt,omitempty"`
	Extend           map[string]*FieldExtractor `json:"extend,omitempty"`
}

// FieldExtractor defines how to extract a specific field from a document
// JavaScript equivalent: { selectors: [...], allowMultiple: bool }.
type FieldExtractor struct {
	Selectors     []interface{} `json:"selectors"`     // Can be string or [string, string] for [selector, attribute]
	AllowMultiple bool          `json:"allowMultiple"` // Allow multiple values
	Format        string        `json:"format"`        // Date format (for date fields)
	Timezone      string        `json:"timezone"`      // Timezone (for date fields)
}

// ContentExtractor defines how to extract and clean main content
// JavaScript equivalent: { selectors: [...], clean: [...], transforms: {...} }.
type ContentExtractor struct {
	*FieldExtractor
	Clean                 []string                     `json:"clean"`                 // Selectors to remove from content
	Transforms            map[string]TransformFunction `json:"transforms"`            // Element transformations
	DisableDefaultCleaner bool                         `json:"disableDefaultCleaner"` // Skip default content cleaning; zero value enables it
}

// TransformFunction represents a function that transforms DOM elements
// JavaScript equivalent: 'selector': $node => { ... } or 'selector': 'tag'.
type TransformFunction interface {
	Transform(selection *goquery.Selection) error
}

// StringTransform is a simple transform that changes tag names
// JavaScript equivalent: 'noscript': 'div'.
type StringTransform struct {
	TargetTag string
}

func (st *StringTransform) Transform(selection *goquery.Selection) error {
	// Convert element to target tag
	html, _ := selection.Html()
	newTag := "<" + st.TargetTag + ">" + html + "</" + st.TargetTag + ">"

	// Replace each element in the selection
	selection.Each(func(i int, s *goquery.Selection) {
		s.ReplaceWithHtml(newTag)
	})

	return nil
}

// FunctionTransform is a custom function transform
// JavaScript equivalent: 'selector': $node => { custom logic }.
type FunctionTransform struct {
	Fn func(*goquery.Selection) error
}

func (ft *FunctionTransform) Transform(selection *goquery.Selection) error {
	return ft.Fn(selection)
}

// ExtractorOptions provides configuration for extraction operations.
type ExtractorOptions struct {
	ContentType string
	Extend      map[string]interface{}
}

// SelectorEntry represents a parsed selector with optional attribute extraction.
type SelectorEntry struct {
	Selector  string
	Attribute string // empty if not extracting attribute
}
