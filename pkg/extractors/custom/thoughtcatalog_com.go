// ABOUTME: ThoughtCatalog.com custom extractor with lifestyle content, writer profiles, and content cleaning
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/thoughtcatalog.com/index.js

package custom

// ThoughtCatalogCustomExtractor provides the custom extraction rules for thoughtcatalog.com
// JavaScript equivalent: export const ThoughtcatalogComExtractor = { ... }
var ThoughtCatalogCustomExtractor = &CustomExtractor{
	Domain: "thoughtcatalog.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.title",
			[]string{"meta[name=\"og:title\"]", "value"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"cite a",
			"div.col-xs-12.article_header div.writer-container.writer-container-inline.writer-no-avatar h4.writer-name",
			"h1.writer-name",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".entry.post",
			},
		},
		
		// No transforms needed for ThoughtCatalog
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			".tc_mark",
			"figcaption",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"article:published_time\"]", "value"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: nil,
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetThoughtCatalogExtractor returns the ThoughtCatalog custom extractor
func GetThoughtCatalogExtractor() *CustomExtractor {
	return ThoughtCatalogCustomExtractor
}