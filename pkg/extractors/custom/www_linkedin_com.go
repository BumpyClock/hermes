// ABOUTME: LinkedIn.com custom extractor with professional content, article format, and author handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.linkedin.com/index.js

package custom

// LinkedInCustomExtractor provides the custom extraction rules for www.linkedin.com
// JavaScript equivalent: export const WwwLinkedinComExtractor = { ... }
var LinkedInCustomExtractor = &CustomExtractor{
	Domain: "www.linkedin.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			".article-title",
			"h1",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".main-author-card h3",
			[]string{"meta[name=\"article:author\"]", "value"},
			".entity-name a[rel=author]",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".article-content__body",
				[]string{"header figure", ".prose"},
				".prose",
			},
		},
		
		// No transforms needed for LinkedIn
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".entity-image",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			".base-main-card__metadata",
			[]string{`time[itemprop="datePublished"]`, "datetime"},
		},
		// Timezone from JavaScript: 'America/Los_Angeles'
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			// enter selectors - empty in JavaScript
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetLinkedInExtractor returns the LinkedIn custom extractor
func GetLinkedInExtractor() *CustomExtractor {
	return LinkedInCustomExtractor
}