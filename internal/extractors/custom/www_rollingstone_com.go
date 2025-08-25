// ABOUTME: Rolling Stone custom extractor for music site with album reviews
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.rollingstone.com/index.js

package custom

// RollingStoneCustomExtractor provides the custom extraction rules for www.rollingstone.com
// JavaScript equivalent: export const WwwRollingstoneComExtractor = { ... }
var RollingStoneCustomExtractor = &CustomExtractor{
	Domain: "www.rollingstone.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.l-article-header__row--title",
			"h1.content-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"a.c-byline__link",
			"a.content-author.tracked-offpage",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
			"time.content-published-date",
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			"h2.l-article-header__row--lead",
			".content-description",
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
				".l-article-content",
				[]string{".lead-container", ".article-content"},
				".article-content",
			},
		},
		
		// No transforms in original JavaScript (empty object)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".c-related-links-wrapper",
			".module-related",
		},
	},
	
	// No selectors in original JavaScript for these fields
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetRollingStoneExtractor returns the Rolling Stone custom extractor
func GetRollingStoneExtractor() *CustomExtractor {
	return RollingStoneCustomExtractor
}