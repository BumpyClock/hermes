// ABOUTME: Engadget custom extractor with complex figure selector patterns
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.engadget.com/index.js

package custom

// WwwEngadgetComExtractor provides the custom extraction rules for www.engadget.com
// JavaScript equivalent: export const WwwEngadgetComExtractor = { ... }
var WwwEngadgetComExtractor = &CustomExtractor{
	Domain: "www.engadget.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"og:title\"]", "value"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			"a.th-meta[data-ylk*=\"subsec:author\"]",
		},
	},
	
	// Engadget stories have publish dates, but the only representation of them on the page
	// is in a format like "2h ago". There are also these tags with blank values:
	// <meta class="swiftype" name="published_at" data-type="date" value="">
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			// Empty in JavaScript
		},
	},
	
	Dek: &FieldExtractor{
		Selectors: []interface{}{
			"div[class*=\"o-title_mark\"] div",
		},
	},
	
	// Engadget stories do have lead images specified by an og:image meta tag, but selecting
	// the value attribute of that tag fails. I believe the "&#x2111;" sequence of characters
	// is triggering this inability to select the attribute value.
	LeadImageURL: &FieldExtractor{
		Selectors: []interface{}{
			// Empty in JavaScript
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				[]interface{}{
					// Some figures will be inside div.article-text, but some header figures/images
					// will not.
					"#page_body figure:not(div.article-text figure)",
					"div.article-text",
				},
			},
		},
		
		// Transform functions (empty in JavaScript)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors (empty in JavaScript)
		Clean: []string{},
	},
}

// GetWwwEngadgetComExtractor returns the Engadget custom extractor
func GetWwwEngadgetComExtractor() *CustomExtractor {
	return WwwEngadgetComExtractor
}