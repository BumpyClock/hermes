// ABOUTME: Engadget custom extractor for www.engadget.com
// ABOUTME: Updated 2025 for new Next.js site structure with data-article-body container

package custom

// WwwEngadgetComExtractor provides the custom extraction rules for www.engadget.com
var WwwEngadgetComExtractor = &CustomExtractor{
	Domain: "www.engadget.com",

	Title: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[property=\"og:title\"]", "content"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []interface{}{
			"a[data-ylk*=\"elm:author\"]",
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[property=\"article:published_time\"]", "content"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[property=\"og:description\"]", "content"},
			[]string{"meta[name=\"description\"]", "content"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[property=\"og:image\"]", "content"},
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
