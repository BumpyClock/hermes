// ABOUTME: Gruene.de (German Green Party) custom extractor with multi-selector content pattern
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.gruene.de/index.js

package custom

// WwwGrueneDeExtractor provides the custom extraction rules for www.gruene.de
// JavaScript equivalent: export const WwwGrueneDeExtractor = { ... }
var WwwGrueneDeExtractor = &CustomExtractor{
	Domain: "www.gruene.de",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"header h1",
		},
	},
	
	// JavaScript: author: null
	Author: nil,
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				// JavaScript comment: selectors: ['section'],
				// JavaScript uses: selectors: [['section header', 'section h2', 'section p', 'section ol']],
				[]string{"section header", "section h2", "section p", "section ol"},
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			"figcaption",
			"p[class]",
		},
	},
	
	// JavaScript: date_published: null
	DatePublished: nil,
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[property=\"og:image\"]", "content"},
		},
	},
	
	// JavaScript: dek: null
	Dek: nil,
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetWwwGrueneDeExtractor returns the Gruene.de custom extractor
func GetWwwGrueneDeExtractor() *CustomExtractor {
	return WwwGrueneDeExtractor
}