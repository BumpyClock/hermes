// ABOUTME: JNSA (Japan Network Security Association) extractor
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.jnsa.org/index.js

package custom

// WwwJnsaOrgExtractor provides the custom extraction rules for www.jnsa.org
// JavaScript equivalent: export const WwwJnsaOrgExtractor = { ... }.
var WwwJnsaOrgExtractor = &CustomExtractor{
	Domain: "www.jnsa.org",

	Title: &FieldExtractor{
		Selectors: []interface{}{
			"#wgtitle h2",
		},
	},

	// Author is null in JavaScript
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:description\"]", "value"},
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
				"#main_area",
			},
		},

		// Clean selectors
		Clean: []string{
			"#pankuzu",
			"#side",
		},
	},
}

// GetWwwJnsaOrgExtractor returns the JNSA custom extractor.
func GetWwwJnsaOrgExtractor() *CustomExtractor {
	return WwwJnsaOrgExtractor
}
