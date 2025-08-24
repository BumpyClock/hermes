// ABOUTME: People.com custom extractor for celebrity and lifestyle content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/people.com/index.js

package custom

// PeopleCustomExtractor provides the custom extraction rules for people.com
// JavaScript equivalent: export const PeopleComExtractor = { ... }
var PeopleCustomExtractor = &CustomExtractor{
	Domain: "people.com",
	
	Title: &FieldExtractor{
		Selectors: []interface{}{
			".article-header h1",
			[]string{"meta[name=\"og:title\"]", "value"},
		},
	},
	
	Author: &FieldExtractor{
		Selectors: []interface{}{
			[]string{"meta[name=\"sailthru.author\"]", "value"},
			"a.author.url.fn",
		},
	},
	
	DatePublished: &FieldExtractor{
		Selectors: []interface{}{
			".mntl-attribution__item-date",
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
			".article-header h2",
		},
	},
	
	Content: &ContentExtractor{
		FieldExtractor: &FieldExtractor{
			Selectors: []interface{}{
				"div[class^=\"loc article-content\"]",
				"div.article-body__inner",
			},
		},
		
		// No transforms in original JavaScript (empty object)
		Transforms: map[string]TransformFunction{},
		
		// Clean selectors - empty array in original JavaScript
		Clean: []string{},
	},
	
	// No selectors in original JavaScript for these fields
	NextPageURL: &FieldExtractor{
		Selectors: []interface{}{},
	},
	
	Excerpt: &FieldExtractor{
		Selectors: []interface{}{},
	},
}

// GetPeopleExtractor returns the People.com custom extractor
func GetPeopleExtractor() *CustomExtractor {
	return PeopleCustomExtractor
}