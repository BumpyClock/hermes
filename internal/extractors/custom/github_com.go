// ABOUTME: GitHub custom extractor with README content and relative-time selectors
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/github.com/index.js

package custom

// GithubComExtractor provides the custom extraction rules for github.com
// JavaScript equivalent: export const GithubComExtractor = { ... }
var GithubComExtractor = &CustomExtractor{
	Domain: "github.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:title\"]", "value"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			// Empty in JavaScript
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"relative-time[datetime]", "datetime"},
			[]string{"span[itemprop=\"dateModified\"] relative-time", "datetime"},
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"description\"]", "value"},
			"span[itemprop=\"about\"]",
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
				[]interface{}{"#readme article"},
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetGithubComExtractor returns the GitHub custom extractor
func GetGithubComExtractor() *CustomExtractor {
	return GithubComExtractor
}