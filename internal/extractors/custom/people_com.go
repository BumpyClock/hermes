// ABOUTME: People.com custom extractor for celebrity and lifestyle content
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/people.com/index.js

package custom

// PeopleCustomExtractor provides the custom extraction rules for people.com
// JavaScript equivalent: export const PeopleComExtractor = { ... }.
var PeopleCustomExtractor = &CustomExtractor{
	Domain: "people.com",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".article-header h1"},
			{Selector: "meta[name=\"og:title\"]", Attribute: "value"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"sailthru.author\"]", Attribute: "value"},
			{Selector: "a.author.url.fn"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".mntl-attribution__item-date"},
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
		},
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".article-header h2"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"div[class^=\"loc article-content\"]"},
			{"div.article-body__inner"},
		},
	},
}

// GetPeopleExtractor returns the People.com custom extractor.
func GetPeopleExtractor() *CustomExtractor {
	return PeopleCustomExtractor
}
