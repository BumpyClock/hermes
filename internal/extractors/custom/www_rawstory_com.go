// ABOUTME: Custom extractor for www.rawstory.com - Progressive news and politics site
// ABOUTME: JavaScript equivalent: src/extractors/custom/www.rawstory.com/index.js WwwRawstoryComExtractor

package custom

// GetWwwRawstoryComExtractor returns the custom extractor for www.rawstory.com.
func GetWwwRawstoryComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "www.rawstory.com",

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:title\"]", Attribute: "value"},
				{Selector: ".blog-title"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "div.main-post-head .social-author__name"},
				{Selector: ".blog-author a:first-of-type"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
				{Selector: ".blog-author a:last-of-type"},
			},
			// Note: timezone: 'EST' is handled by date cleaner in Go version
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".post-body"},
				{".blog-content"},
			},

			Transforms: map[string]TransformFunction{
				// No transforms in JavaScript version
			},

			Clean: []string{
				// No clean selectors in JavaScript version
			},
		},
	}
}
