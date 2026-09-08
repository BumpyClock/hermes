// ABOUTME: Engadget custom extractor for www.engadget.com
// ABOUTME: Updated 2025 for new Next.js site structure with data-article-body container
// ABOUTME: Note: Hermes normalizes meta tags (property->name, content->value)

package custom

// WwwEngadgetComExtractor provides the custom extraction rules for www.engadget.com.
var WwwEngadgetComExtractor = &CustomExtractor{
	Domain: "www.engadget.com",

	// Note: Hermes normalizes meta tags: property->name, content->value
	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:title\"]", Attribute: "value"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "a[data-ylk*=\"elm:author\"]"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:description\"]", Attribute: "value"},
			{Selector: "meta[name=\"description\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div[data-article-body=\"true\"]"},
			{"article"},
		},

		// Clean out ads, commerce modules, and non-content elements
		Clean: []string{
			".productModule",
			".commerce",
			"[class*=\"Advertisement\"]",
			"nav",
			"footer",
		},
	},
}

// GetWwwEngadgetComExtractor returns the Engadget custom extractor.
func GetWwwEngadgetComExtractor() *CustomExtractor {
	return WwwEngadgetComExtractor
}
