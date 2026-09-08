// ABOUTME: E! Online custom extractor for entertainment industry content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.eonline.com/index.js

package custom

// EOnlineCustomExtractor provides the custom extraction rules for www.eonline.com
// JavaScript equivalent: export const WwwEonlineComExtractor = { ... }.
var EOnlineCustomExtractor = &CustomExtractor{
	Domain: "www.eonline.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1.article-detail__title"},
			{Selector: "h1.article__title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".article-detail__meta__author"},
			{Selector: ".entry-meta__author a"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
			{Selector: "meta[itemprop=\"datePublished\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".article-detail__main-content section"},
			{".post-content section, .post-content div.post-content__image"},
		},

		// Transform functions for E! Online-specific content
		Transforms: map[string]TransformFunction{
			"div.post-content__image":                 &StringTransform{TargetTag: "figure"},
			"div.post-content__image .image__credits": &StringTransform{TargetTag: "figcaption"},
		},
	},
}

// GetEOnlineExtractor returns the E! Online custom extractor.
func GetEOnlineExtractor() *CustomExtractor {
	return EOnlineCustomExtractor
}
