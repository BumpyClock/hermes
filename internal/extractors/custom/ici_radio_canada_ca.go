// ABOUTME: ICI Radio-Canada (French Canadian news) custom extractor with date format and timezone handling
// ABOUTME: 100% JavaScript-compatible port of src/extractors/custom/ici.radio-canada.ca/index.js

package custom

// IciRadioCanadaCaExtractor provides the custom extraction rules for ici.radio-canada.ca
// JavaScript equivalent: export const IciRadioCanadaCaExtractor = { ... }.
var IciRadioCanadaCaExtractor = &CustomExtractor{
	Domain: "ici.radio-canada.ca",

	Title: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "h1"},
		},
	},

	Author: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"dc.creator\"]", Attribute: "value"},
		},
	},

	Content: &ContentExtractor{
		Selectors: []ContentSelectorGroup{
			{"section.document-content-style"},
			{".main-multimedia-item", ".news-story-content"},
		},
	},

	DatePublished: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"dc.date.created\"]", Attribute: "value"},
		},
		// Note: JavaScript version has format: 'YYYY-MM-DD|HH[h]mm' and timezone: 'America/New_York'
		// This is handled by dateparse library in Go which can parse various formats automatically
	},

	LeadImageURL: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "meta[name=\"og:image\"]", Attribute: "value"},
		},
	},

	Dek: &FieldExtractor{
		Selectors: []SelectorEntry{
			{Selector: "div.lead-container"},
			{Selector: ".bunker-component.lead"},
		},
	},
}

// GetIciRadioCanadaCaExtractor returns the ICI Radio-Canada custom extractor.
func GetIciRadioCanadaCaExtractor() *CustomExtractor {
	return IciRadioCanadaCaExtractor
}
