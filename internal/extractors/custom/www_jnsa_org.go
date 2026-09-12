// ABOUTME: JNSA (Japan Network Security Association) extractor
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.jnsa.org/index.js

package custom

// WwwJnsaOrgExtractor provides the custom extraction rules for www.jnsa.org
// JavaScript equivalent: export const WwwJnsaOrgExtractor = { ... }.
var WwwJnsaOrgExtractor = &CustomExtractor{
	Domain: "www.jnsa.org",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "#wgtitle h2"},
		},
	},

	// Author is null in JavaScript
	Excerpt: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:description\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"#main_area"},
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
