package custom

// GeniusCustomExtractor provides the custom extraction rules for genius.com
// JavaScript equivalent: export const GeniusComExtractor = { ... }.
var GeniusCustomExtractor = &CustomExtractor{
	Domain: "genius.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h2 a"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".lyrics"},
		},
	},
}

// GetGeniusExtractor returns the Genius custom extractor.
func GetGeniusExtractor() *CustomExtractor {
	return GeniusCustomExtractor
}
