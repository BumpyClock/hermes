// ABOUTME: Custom extractor for gothamist.com - NYC local news with multi-city support
// ABOUTME: JavaScript equivalent: src/extractors/custom/gothamist.com/index.js GothamistComExtractor

package custom

// GetGothamistComExtractor returns the custom extractor for gothamist.com and related city sites.
func GetGothamistComExtractor() *CustomExtractor {
	return &CustomExtractor{
		Domain: "gothamist.com",

		// SupportedDomains from JavaScript version
		SupportedDomains: []string{
			"chicagoist.com",
			"laist.com",
			"sfist.com",
			"shanghaiist.com",
			"dcist.com",
		},

		Title: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "h1"},
				{Selector: ".entry-header h1"},
			},
		},

		Author: &FieldExtractor{
			Selectors: []SelectorEntry{
				// Comment from JavaScript: "There are multiple article-metadata and byline-author classes, but the main article's is the 3rd child of the l-container class"
				{Selector: ".article-metadata:nth-child(3) .byline-author"},
				{Selector: ".author"},
			},
		},

		DatePublished: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
				{Selector: "abbr"},
				{Selector: "abbr.published"},
			},
			// Note: timezone: 'America/New_York' is handled by date cleaner in Go version
		},

		LeadImageURL: &FieldExtractor{
			Selectors: []SelectorEntry{
				{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
			},
		},

		Content: &ContentExtractor{
			Selectors: []ContentSelectorGroup{
				{".article-body"},
				{".entry-body"},
			},

			Transforms: map[string]TransformFunction{
				// String transforms for image classes to figures
				"div.image-none":  &StringTransform{TargetTag: "figure"},
				".image-none i":   &StringTransform{TargetTag: "figcaption"},
				"div.image-left":  &StringTransform{TargetTag: "figure"},
				".image-left i":   &StringTransform{TargetTag: "figcaption"},
				"div.image-right": &StringTransform{TargetTag: "figure"},
				".image-right i":  &StringTransform{TargetTag: "figcaption"},
			},

			Clean: []string{
				".image-none br",
				".image-left br",
				".image-right br",
				".galleryEase",
			},
		},
	}
}
