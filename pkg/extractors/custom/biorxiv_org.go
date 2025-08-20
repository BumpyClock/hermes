// ABOUTME: BioRxiv research preprint server extractor for academic papers
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/biorxiv.org/index.js

package custom

// BiorxivOrgExtractor provides the custom extraction rules for biorxiv.org
// JavaScript equivalent: export const BiorxivOrgExtractor = { ... }
var BiorxivOrgExtractor = &CustomExtractor{
	Domain: "biorxiv.org",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			"h1#page-title",
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"div.highwire-citation-biorxiv-article-top > div.highwire-cite-authors",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div#abstract-1",
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetBiorxivOrgExtractor returns the BioRxiv custom extractor
func GetBiorxivOrgExtractor() *CustomExtractor {
	return BiorxivOrgExtractor
}