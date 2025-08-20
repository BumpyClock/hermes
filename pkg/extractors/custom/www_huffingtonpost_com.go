// ABOUTME: HuffingtonPost.com custom extractor with news article support and content cleaning
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.huffingtonpost.com/index.js

package custom

// HuffingtonPostCustomExtractor provides the custom extraction rules for www.huffingtonpost.com
// JavaScript equivalent: export const WwwHuffingtonpostComExtractor = { ... }
var HuffingtonPostCustomExtractor = &CustomExtractor{
	Domain: "www.huffingtonpost.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.headline__title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"span.author-card__details__name",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div.entry__body",
			},
			DefaultCleaner: false,
		},
		
		// No transforms needed for HuffPost
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".pull-quote",
			".tag-cloud",
			".embed-asset",
			".below-entry",
			".entry-corrections",
			"#suggested-story",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:modified_time\"]", "value"},
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			"h2.headline__subtitle",
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetHuffingtonPostExtractor returns the HuffingtonPost custom extractor
func GetHuffingtonPostExtractor() *CustomExtractor {
	return HuffingtonPostCustomExtractor
}