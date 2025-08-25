// ABOUTME: Le Monde (www.lemonde.fr) custom extractor for French news content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.lemonde.fr/index.js

package custom

// WwwLemondeFrExtractor provides the custom extraction rules for www.lemonde.fr
// JavaScript equivalent: export const WwwLemondeFrExtractor = { ... }
var WwwLemondeFrExtractor = &CustomExtractor{
	Domain: "www.lemonde.fr",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1.article__title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			".author__name",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				".article__content",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - remove unwanted elements
		Clean: []string{
			"figcaption",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:article:published_time\"]", "value"},
		},
	},
	
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:image\"]", "value"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			".article__desc",
		},
	},
	
	NextPageURL: nil,
	
	Excerpt: nil,
}

// GetWwwLemondeFrExtractor returns the Le Monde custom extractor
func GetWwwLemondeFrExtractor() *CustomExtractor {
	return WwwLemondeFrExtractor
}