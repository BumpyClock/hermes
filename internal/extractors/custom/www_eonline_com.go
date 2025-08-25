// ABOUTME: E! Online custom extractor for entertainment industry content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.eonline.com/index.js

package custom

// EOnlineCustomExtractor provides the custom extraction rules for www.eonline.com
// JavaScript equivalent: export const WwwEonlineComExtractor = { ... }
var EOnlineCustomExtractor = &CustomExtractor{
	Domain: "www.eonline.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.article-detail__title",
			"h1.article__title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".article-detail__meta__author",
			".entry-meta__author a",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
			[]string{"meta[itemprop=\"datePublished\"]", "value"},
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
				".article-detail__main-content section",
				".post-content section, .post-content div.post-content__image",
			},
		},
		
		// Transform functions for E! Online-specific content
		Transforms: map[string]TransformFunction{
			"div.post-content__image": &StringTransform{TargetTag: "figure"},
			"div.post-content__image .image__credits": &StringTransform{TargetTag: "figcaption"},
		},
		
		// Clean selectors - empty array in original JavaScript
		Clean: []string{},
	},
	
	// No selectors in original JavaScript for these fields
	Dek: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetEOnlineExtractor returns the E! Online custom extractor
func GetEOnlineExtractor() *CustomExtractor {
	return EOnlineCustomExtractor
}