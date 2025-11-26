// ABOUTME: Engadget custom extractor for www.engadget.com
// ABOUTME: Updated 2025 for new Next.js site structure with data-article-body container
// ABOUTME: Note: Hermes normalizes meta tags (property->name, content->value)

package custom

// WwwEngadgetComExtractor provides the custom extraction rules for www.engadget.com
var WwwEngadgetComExtractor = &CustomExtractor{
	Domain: "www.engadget.com",

	// Note: Hermes normalizes meta tags: property->name, content->value
	Title: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:title\"]", "value"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []interface{}{
			"a[data-ylk*=\"elm:author\"]",
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:description\"]", "value"},
			[]string{"meta[name=\"description\"]", "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},

	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div[data-article-body=\"true\"]",
				"article",
			},
		},

		// Clean out ads, commerce modules, and non-content elements
		Clean: []string{
			".productModule",
			".commerce",
			"[class*=\"Advertisement\"]",
			"nav",
			"footer",
		},

		Transforms: map[string]TransformFunction{},
	},
}

// GetWwwEngadgetComExtractor returns the Engadget custom extractor
func GetWwwEngadgetComExtractor() *CustomExtractor {
	return WwwEngadgetComExtractor
}
