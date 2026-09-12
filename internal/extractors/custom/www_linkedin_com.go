// ABOUTME: LinkedIn.com custom extractor with professional content, article format, and author handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.linkedin.com/index.js

package custom

// LinkedInCustomExtractor provides the custom extraction rules for www.linkedin.com
// JavaScript equivalent: export const WwwLinkedinComExtractor = { ... }.
var LinkedInCustomExtractor = &CustomExtractor{
	Domain: "www.linkedin.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".article-title"},
			{Selector: "h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".main-author-card h3"},
			{Selector: "meta[name=\"article:author\"]", Attribute: "value"},
			{Selector: ".entity-name a[rel=author]"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".article-content__body"},
			{"header figure", ".prose"},
			{".prose"},
		},

		// Clean selectors - remove unwanted elements
		Clean: []string{
			".entity-image",
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".base-main-card__metadata"},
			{Selector: `time[itemprop="datePublished"]`, Attribute: "datetime"},
		},
		// Timezone from JavaScript: 'America/Los_Angeles'
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},
}

// GetLinkedInExtractor returns the LinkedIn custom extractor.
func GetLinkedInExtractor() *CustomExtractor {
	return LinkedInCustomExtractor
}
