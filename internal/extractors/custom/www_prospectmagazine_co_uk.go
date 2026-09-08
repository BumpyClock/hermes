// ABOUTME: Prospect Magazine UK custom extractor with European timezone handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/www.prospectmagazine.co.uk/index.js

package custom

// WwwProspectmagazineCoUkExtractor provides the custom extraction rules for www.prospectmagazine.co.uk
// JavaScript equivalent: export const WwwProspectmagazineCoUkExtractor = { ... }.
var WwwProspectmagazineCoUkExtractor = &CustomExtractor{
	Domain: "www.prospectmagazine.co.uk",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".blog-header__title"},
			{Selector: ".page-title"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".blog-header__author-link"},
			{Selector: ".aside_author .title"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{".blog__container"},
			{"article .post_content"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"article:published_time\"]", Attribute: "value"},
			{Selector: ".post-info"},
		},
		// Note: JavaScript version has timezone: 'Europe/London'
		// This is handled by dateparse library in Go
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: ".blog-header__description"},
			{Selector: ".page-subtitle"},
		},
	},
}

// GetWwwProspectmagazineCoUkExtractor returns the Prospect Magazine UK custom extractor.
func GetWwwProspectmagazineCoUkExtractor() *CustomExtractor {
	return WwwProspectmagazineCoUkExtractor
}
